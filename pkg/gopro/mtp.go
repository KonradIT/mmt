package gopro

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/konradit/gowpd"
	"github.com/konradit/mmt/pkg/utils"
)

var SYSTEM_VOLUME_INFO = "System Volume Information"

var CameraSeries = map[GoProType][]string{
	V2: {
		"HERO9",
		"HERO8",
		"HERO7",
		"HERO6",
	},
	MAX: {
		"GoPro",
	},
	V1: {
		"HERO5",
		"HERO", // HERO Session, HERO LCD, etc...
		"HERO+",
		"HERO4",
		"HERO3",
	},
}

func listMtpFiles(d *gowpd.Device, id string, curPath string, clean bool, list map[string]*gowpd.Object) int {
	objs, _ := d.GetChildObjects(id)
	n := len(objs)
	for _, o := range objs {
		if o == nil {
			continue
		}
		if o.Name == SYSTEM_VOLUME_INFO {
			n--
			continue
		} else if strings.HasPrefix(o.Name, ".trashed-") {
			n--
			continue
		}
		rel := filepath.Join(curPath, o.Name)
		list[rel] = o
		if o.IsDir {
			o.ChildCount = listMtpFiles(d, o.Id, rel, clean, list)
			if clean && o.ChildCount == 0 {
				delete(list, rel)
				n--
			}
		} else {
			o.ChildCount = -1
			if clean && o.Size == 0 {
				delete(list, rel)
				n--
			}
		}
	}
	return n
}

func ListMtpFiles(d *gowpd.Device, path string, clean bool) (list map[string]*gowpd.Object) {
	list = make(map[string]*gowpd.Object)
	obj := d.FindObject(path)
	if obj == nil || !obj.IsDir {
		return
	}
	listMtpFiles(d, obj.Id, "", clean, list)
	return
}

func ImportViaMTP(in, output string, sortoptions SortOptions) (*utils.Result, error) {
	err := gowpd.Init()
	defer gowpd.Destroy()
	if err != nil {
		return nil, err
	}
	n := gowpd.GetDeviceCount()
	deviceChosenIndex := -1
	for i := 0; i < n; i++ {
		if strings.ToLower(gowpd.GetDeviceName(i)) == strings.ToLower(in) {
			deviceChosenIndex = i
		}
	}

	if deviceChosenIndex == -1 {
		return nil, errors.New("Device with name `" + in + "` not found.")
	}
	goproType := UNKNOWN

	rootname := strings.Split(gowpd.GetDeviceName(deviceChosenIndex), " ")[0]
	for familyname, cameras := range CameraSeries {
		for _, camera := range cameras {
			if strings.Contains(camera, rootname) {
				goproType = familyname
				break
			}
		}
	}
	if goproType == UNKNOWN {
		return nil, errors.New("GoPro type not identified.")
	}

	d, err := gowpd.ChooseDevice(deviceChosenIndex)
	var result utils.Result
	if err == nil {

		o := d.FindObject(gowpd.PathSeparator) //find root object
		if o != nil {

			r := ListMtpFiles(d, "", true)
			for _, t := range r {
				if t.IsDir {
					continue
				}

				if t.Name == string(GetStarted) {
					continue
				}
				timestamp := time.Unix(t.CreatedAtTime, 0)
				mediaDate := timestamp.Format("02-01-2006")
				if strings.Contains(sortoptions.DateFormat, "year") && strings.Contains(sortoptions.DateFormat, "month") && strings.Contains(sortoptions.DateFormat, "day") {

					mediaDate = timestamp.Format(replacer.Replace(sortoptions.DateFormat))
				}

				if len(sortoptions.DateRange) == 2 {

					start := sortoptions.DateRange[0]
					end := sortoptions.DateRange[1]
					if timestamp.Before(start) {
						continue
					}
					if timestamp.After(end) {
						continue
					}

				}

				dayFolder := filepath.Join(output, mediaDate)
				if _, err := os.Stat(dayFolder); os.IsNotExist(err) {
					os.Mkdir(dayFolder, 0755)
				}

				if sortoptions.ByCamera {
					if _, err := os.Stat(filepath.Join(dayFolder, in)); os.IsNotExist(err) {
						os.Mkdir(filepath.Join(dayFolder, in), 0755)
					}
					dayFolder = filepath.Join(dayFolder, in)
				}

				for _, fileTypeMatch := range FileTypeMatches[goproType] {

					if fileTypeMatch.Regex.MatchString(t.Name) {
						switch goproType {

						case V2:
							switch fileTypeMatch.Type {
							case Video:
								x := t.Name
								filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
								color.Green(">>> %s", x)

								if _, err := os.Stat(filepath.Join(dayFolder, "videos")); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "videos"), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}

								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "videos", filename), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}
							case Photo:
								if _, err := os.Stat(filepath.Join(dayFolder, "photos")); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "photos"), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}

								color.Green(">>> %s", t.Name)

								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "photos", t.Name), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}
							case Multishot:
								filebaseroot := t.Name[:4]
								if _, err := os.Stat(filepath.Join(dayFolder, "multishot", filebaseroot)); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "multishot", filebaseroot), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}

								color.Green(">>> %s/%s", filebaseroot, t.Name)

								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "multishot", filebaseroot, t.Name), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}
							case Thumbnail:
								if !sortoptions.SkipAuxiliaryFiles {
									if _, err := os.Stat(filepath.Join(dayFolder, "videos/proxy")); os.IsNotExist(err) {
										err = os.MkdirAll(filepath.Join(dayFolder, "videos/proxy"), 0755)
										if err != nil {
											log.Fatal(err.Error())
										}
									}

									x := t.Name
									filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
									_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "videos/proxy", filename), t)
									if err != nil {
										result.Errors = append(result.Errors, err)
										result.FilesNotImported = append(result.FilesNotImported, filename)
									} else {
										result.FilesImported += 1
									}
								}
							case LowResolutionVideo:
								if !sortoptions.SkipAuxiliaryFiles {
									if _, err := os.Stat(filepath.Join(dayFolder, "videos/proxy")); os.IsNotExist(err) {
										err = os.MkdirAll(filepath.Join(dayFolder, "videos/proxy"), 0755)
										if err != nil {
											log.Fatal(err.Error())
										}
									}

									x := t.Name
									filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
									_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "videos/proxy", filename), t)
									if err != nil {
										result.Errors = append(result.Errors, err)
										result.FilesNotImported = append(result.FilesNotImported, filename)
									} else {
										result.FilesImported += 1
									}
								}
							case RawPhoto:
								if _, err := os.Stat(filepath.Join(dayFolder, "photos/raw")); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "photos/raw"), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}

								color.Green(">>> %s", t.Name)
								// convert to DNG here
								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "photos/raw", t.Name), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}
							default:
								color.Red("Unsupported file %s", t.Name)
								result.Errors = append(result.Errors, errors.New("Media format unrecognized"))
								result.FilesNotImported = append(result.FilesNotImported, t.Name)
							}
						case V1:
							switch fileTypeMatch.Type {
							case Video:
								x := t.Name

								chaptered := regexp.MustCompile(`GP\d+.MP4`)
								if chaptered.MatchString(t.Name) {
									x = fmt.Sprintf("GOPR%s%s.%s", x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
								}
								color.Green(">>> %s", x)

								if _, err := os.Stat(filepath.Join(dayFolder, "videos")); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "videos"), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}

								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "videos", x), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, x)
								} else {
									result.FilesImported += 1
								}
							case ChapteredVideo:
								x := t.Name
								name := fmt.Sprintf("GOPR%s%s.%s", x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])

								color.Green(">>> %s", x)

								if _, err := os.Stat(filepath.Join(dayFolder, "videos")); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "videos"), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}
								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "videos", name), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}
							case Photo:
								if _, err := os.Stat(filepath.Join(dayFolder, "photos")); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "photos"), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}

								color.Green(">>> %s", t.Name)

								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "photos", t.Name), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}

							case LowResolutionVideo:
								if !sortoptions.SkipAuxiliaryFiles {
									if _, err := os.Stat(filepath.Join(dayFolder, "videos/proxy")); os.IsNotExist(err) {
										err = os.MkdirAll(filepath.Join(dayFolder, "videos/proxy"), 0755)
										if err != nil {
											log.Fatal(err.Error())
										}
									}

									_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "videos/proxy", t.Name), t)
									if err != nil {
										result.Errors = append(result.Errors, err)
										result.FilesNotImported = append(result.FilesNotImported, t.Name)
									} else {
										result.FilesImported += 1
									}
								}
							case Thumbnail:
								if !sortoptions.SkipAuxiliaryFiles {
									if _, err := os.Stat(filepath.Join(dayFolder, "videos/proxy")); os.IsNotExist(err) {
										err = os.MkdirAll(filepath.Join(dayFolder, "videos/proxy"), 0755)
										if err != nil {
											log.Fatal(err.Error())
										}
									}

									_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "videos/proxy", t.Name), t)
									if err != nil {
										result.Errors = append(result.Errors, err)
										result.FilesNotImported = append(result.FilesNotImported, t.Name)
									} else {
										result.FilesImported += 1
									}
								}
							case Multishot:
								filebaseroot := t.Name[:4]
								if _, err := os.Stat(filepath.Join(dayFolder, "multishot", filebaseroot)); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "multishot", filebaseroot), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}

								color.Green(">>> %s/%s", filebaseroot, t.Name)

								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "multishot", filebaseroot, t.Name), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}

							case RawPhoto:
								if _, err := os.Stat(filepath.Join(dayFolder, "photos/raw")); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "photos/raw"), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}

								color.Green(">>> %s", t.Name)
								// convert to DNG here
								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "photos/raw", t.Name), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}
							default:
								color.Red("Unsupported file %s", t.Name)
								result.Errors = append(result.Errors, errors.New("Media format unrecognized"))
								result.FilesNotImported = append(result.FilesNotImported, t.Name)
							}
						case MAX:
							switch fileTypeMatch.Type {
							case Video:
								x := t.Name

								filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
								color.Green(">>> %s", x)

								foldersNeeded := []string{"videos/360", "videos/heromode"}
								for _, fn := range foldersNeeded {
									if _, err := os.Stat(filepath.Join(dayFolder, fn)); os.IsNotExist(err) {
										err = os.MkdirAll(filepath.Join(dayFolder, fn), 0755)
										if err != nil {
											log.Fatal(err.Error())
										}
									}
								}

								dest := foldersNeeded[1]
								if !fileTypeMatch.HeroMode {
									dest = foldersNeeded[0]
								}
								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, dest, filename), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}
							case Photo:
								foldersNeeded := []string{"photos/360", "photos/heromode"}
								for _, fn := range foldersNeeded {
									if _, err := os.Stat(filepath.Join(dayFolder, fn)); os.IsNotExist(err) {
										err = os.MkdirAll(filepath.Join(dayFolder, fn), 0755)
										if err != nil {
											log.Fatal(err.Error())
										}
									}
								}

								dest := foldersNeeded[1]
								if !fileTypeMatch.HeroMode {
									dest = foldersNeeded[0]
								}
								color.Green(">>> %s", t.Name)

								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, dest, t.Name), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}
							case PowerPano:
								if _, err := os.Stat(filepath.Join(dayFolder, "photos/powerpano")); os.IsNotExist(err) {
									err = os.MkdirAll(filepath.Join(dayFolder, "photos/powerpano"), 0755)
									if err != nil {
										log.Fatal(err.Error())
									}
								}

								color.Green(">>> %s", t.Name)

								_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, "photos/powerpano", t.Name), t)
								if err != nil {
									result.Errors = append(result.Errors, err)
									result.FilesNotImported = append(result.FilesNotImported, t.Name)
								} else {
									result.FilesImported += 1
								}
							case LowResolutionVideo:
								if !sortoptions.SkipAuxiliaryFiles {
									foldersNeeded := []string{"videos/proxy/heromode", "videos/proxy/360"}
									for _, fn := range foldersNeeded {
										if _, err := os.Stat(filepath.Join(dayFolder, fn)); os.IsNotExist(err) {
											err = os.MkdirAll(filepath.Join(dayFolder, fn), 0755)
											if err != nil {
												log.Fatal(err.Error())
											}
										}
									}
									dest := foldersNeeded[1]
									if fileTypeMatch.HeroMode {
										dest = foldersNeeded[0]
									}
									x := t.Name

									filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
									_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, dest, filename), t)
									if err != nil {
										result.Errors = append(result.Errors, err)
										result.FilesNotImported = append(result.FilesNotImported, t.Name)
									} else {
										result.FilesImported += 1
									}
								}
							case Thumbnail:
								if !sortoptions.SkipAuxiliaryFiles {
									foldersNeeded := []string{"videos/thumbnails/heromode", "videos/thumbnails/360"}
									for _, fn := range foldersNeeded {
										if _, err := os.Stat(filepath.Join(dayFolder, fn)); os.IsNotExist(err) {
											err = os.MkdirAll(filepath.Join(dayFolder, fn), 0755)
											if err != nil {
												log.Fatal(err.Error())
											}
										}
									}
									dest := foldersNeeded[1]
									if fileTypeMatch.HeroMode {
										dest = foldersNeeded[0]
									}
									x := t.Name

									filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
									_, err := d.CopyObjectFromDevice(filepath.Join(dayFolder, dest, filename), t)
									if err != nil {
										result.Errors = append(result.Errors, err)
										result.FilesNotImported = append(result.FilesNotImported, t.Name)
									} else {
										result.FilesImported += 1
									}
								}
							default:
								color.Red("Unsupported file %s", t.Name)
								result.Errors = append(result.Errors, errors.New("Unsupported file "+t.Name))
								result.FilesNotImported = append(result.FilesNotImported, t.Name)
							}
						default:
							return nil, errors.New("Camera not recognized.")
						}
					}
				}
			}
		}
	}
	return &result, nil
}
