package nova

import (
	"errors"
	"io/fs"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/cred"
	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/fileIO"
	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/helpers"
)

var (
	ErrMalformedRow CustomSkinError = CustomSkinError{"malformed row"}
)

type CustomSkinError struct {
	name string
}

func (cse CustomSkinError) Error() string {
	return cse.name
}

// ~~~~~~~~~~~~~~~~~ CustomSkin
type CustomSkin struct {
	pictures []fileIO.File
	credit   cred.CreditType

	name        string
	body        fileIO.File
	forceArmour fileIO.File
	drone       fileIO.File
	angle       string
	distance    string

	zip fileIO.ZipFile
}

func NewCustomSkin(name, angle, distance string) (cs *CustomSkin) {
	cs = &CustomSkin{name: name, angle: angle, distance: distance}

	return cs
}

func (cs *CustomSkin) addBody(f fileIO.File) *CustomSkin {
	cs.body = f
	return cs
}

func (cs *CustomSkin) addForceA(s fileIO.File) *CustomSkin {
	cs.forceArmour = s
	return cs
}

func (cs *CustomSkin) addDrone(f fileIO.File) *CustomSkin {
	cs.drone = f
	return cs
}

func (cs *CustomSkin) addCredits(c cred.CreditType) {
	cs.credit = c
}

func (cs *CustomSkin) addMedia(f fileIO.File) {
	cs.pictures = append(cs.pictures, f)
}

func (cs *CustomSkin) HasZip() bool {
	return reflect.DeepEqual(&cs.zip, (&fileIO.ZipFile{}))
}

// TODO This should use the fs.DirEntires to generate a zip file for the individual skin
func (cs *CustomSkin) generateZipFile() {
	cs.zip = *fileIO.NewZipFile(cs.name)

	body, fA, drone := cs.getBody_FA_Drone()

	cs.zip.AddZipFile(body, []byte(body))
	cs.zip.AddZipFile(fA, []byte(fA))
	cs.zip.AddZipFile(drone, []byte(drone))

	return
}

func (cs *CustomSkin) String() string {
	return cs.name
}

func OpenCustomSkin(d fs.DirEntry) fileIO.File {
	return fileIO.OpenFile(d)
}

func EmptyCustomSkin() *CustomSkin {
	return &CustomSkin{}
}

func CSVLineToCustomSkin(s string, custom_skin_dir []os.DirEntry, reqLength int) (cs *CustomSkin, err error) {
	ss := strings.Split(s, ",")

	helpers.Print(len(ss), reqLength)

	if len(ss) != reqLength {
		return EmptyCustomSkin(), ErrMalformedRow
	}

	bodyS, forceArmourS, droneS := ss[1], ss[2], ss[3]

	body, _ := fileIn(bodyS, custom_skin_dir)
	forceArmour, _ := fileIn(forceArmourS, custom_skin_dir)
	drone, _ := fileIn(droneS, custom_skin_dir)

	cs = NewCustomSkin(ss[0], ss[4], ss[5]).addBody(body).addForceA(forceArmour).addDrone(drone)

	cs.generateZipFile()

	return
}

// TODO replace this with the SearchWithFunc when you update the helpers library version used
func fileIn(s string, arr []os.DirEntry) (f fileIO.File, err error) {
	f = *fileIO.EmptyFile()
	err = errors.New("file not found")

	if reflect.DeepEqual("", s) {
		return f, errors.New("empty file")
	}

	for _, v := range arr {
		if reflect.DeepEqual(s, v.Name()) {
			return fileIO.OpenFile(v), nil
		}
	}

	return
}

func emptyOSFile() os.File {
	return os.File{}
}

func (cs CustomSkin) getBody_FA_Drone() (body, fA, drone string) {
	body, fA, drone = "", "", ""

	if !fileIO.FilesEqual(cs.body, *fileIO.EmptyFile()) {
		body = cs.body.Name()
	}

	if !fileIO.FilesEqual(cs.forceArmour, *fileIO.EmptyFile()) {
		fA = cs.forceArmour.Name()
	}

	if !fileIO.FilesEqual(cs.drone, *fileIO.EmptyFile()) {
		drone = cs.drone.Name()
	}

	return
}

func (cs *CustomSkin) toCSVLine() string {
	body, fA, drone := cs.getBody_FA_Drone()
	return format("%s,%s,%s,%s,%s,%s", cs.name, body, fA, drone, cs.getAngle(), cs.getDistance())
}

func (c *CustomSkin) getAngle() string {
	// try to convert s to an integer, if it fails, return nothing
	_, err := strconv.Atoi(c.angle)
	if err != nil {
		return ""
	} else {
		return c.angle
	}
}

func (c *CustomSkin) getDistance() string {
	// try to convert to an integer
	_, err := strconv.Atoi(c.distance)
	if err != nil {
		return ""
	} else {
		return c.distance
	}
}

func (c *CustomSkin) FormatCredits() string {
	if c.credit == nil {
		return ""
	}
	return fileIO.ConstructMarkDownLink(false, c.credit.ConstructName(), c.credit.ConstructLink())
}

// returns a list of CustomSkins based on whats in the custom_skins folder
func GetCustomSkins(custom_skin_dir []fs.DirEntry) (skins []CustomSkin) {
	skinsData := fileIO.ReadCSV(inSkinsFolder("custom_skins"))
	credits := skinsData.GetIndexOfColumn("credit")

	discordUIDs := getDiscordUIDs()
	infoMaps := []map[string]string{discordUIDs}
	mapType := []cred.CreditSource{cred.Discord}

	reqLength := skinsData.NumHeaders()
	skins = make([]CustomSkin, 0, skinsData.Rows())

	for row := range skinsData.Rows() {
		s := skinsData.GetRow(row)
		helpers.Print("[", s, "]")
		skin, err := CSVLineToCustomSkin(s, custom_skin_dir, reqLength)
		if err != nil {
			// printf("malformed csv, required length: %d, length: %d, %s,", reqLength, len(s), s)
			helpers.Print("Get Custom Skin Error", s)
			if !errors.Is(err, ErrMalformedRow) {
				helpers.Handle(err)
			}

			continue
		}
		// print(i, v, body, forces, drones)

		credit := skinsData.GetCell(row, credits)
		creditInfo, creditType := assignCredits(credit, infoMaps, mapType)

		if creditType != cred.Default {
			skin.addCredits(cred.NewCredit(credit, creditInfo, creditType))
		}

		// construct zip file for skin

		skins = append(skins, *skin)

		// printf("appropriate length: %d, %s", len(v), skin)

	}

	return
}

func assignCredits(credit string, creditInfoMaps []map[string]string, mapTypes []cred.CreditSource) (creditInfo string, creditType cred.CreditSource) {
	// assign credit type based on credit info
	for i, m := range creditInfoMaps {
		temp, exists := m[credit]
		if exists {
			creditInfo = temp
			creditType = mapTypes[i]
			return
		}
	}

	creditType = cred.Default
	return
}

func getDiscordUIDs() map[string]string {
	discordCreditData := fileIO.ReadCSV(inAssetsFolder("DISCORD_UIDS"))
	fileContents := discordCreditData.GetContents()

	uidMap := make(map[string]string, discordCreditData.Rows())

	for _, row := range fileContents {
		discordName := row[0]
		UID := row[1]
		uidMap[discordName] = UID
	}

	return uidMap
}
