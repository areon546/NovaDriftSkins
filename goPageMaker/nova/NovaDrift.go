package nova

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/cred"
	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/fileIO"
	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/helpers"
)

// ~~~~~~~~~~~~~~~~~ CustomSkin
type CustomSkin struct {
	pictures []fileIO.File
	credit   cred.CreditType

	name        string
	body        string
	forceArmour string
	drone       string
	angle       string
	distance    string
}

func NewCustomSkin(name, angle, distance string) *CustomSkin {
	return &CustomSkin{name: name, angle: angle, distance: distance}
}

func (c *CustomSkin) addBody(b string) *CustomSkin {
	c.body = b
	return c
}

func (c *CustomSkin) addForceA(s string) *CustomSkin {
	c.forceArmour = s
	return c
}

func (c *CustomSkin) addDrone(s string) *CustomSkin {
	c.drone = s
	return c
}

func (cs *CustomSkin) addCredits(c cred.CreditType) {
	cs.credit = c
}

func (c *CustomSkin) String() string {
	return c.name
}

func convertCSVLineToCustomSkin(s string) *CustomSkin {
	ss := strings.Split(s, ",")
	c := CustomSkin{name: ss[0], body: ss[1], forceArmour: ss[2], drone: ss[3], angle: ss[4], distance: ss[5]}
	return &c
}

func (c *CustomSkin) toCSVLine() string {
	return format("%s,%s,%s,%s,%s,%s", c.name, c.body, c.forceArmour, c.drone, c.getAngle(), c.getDistance())
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
func GetCustomSkins() (skins []CustomSkin) {
	skinsData := fileIO.ReadCSV(skinFolder() + "custom_skins")
	names := skinsData.GetIndexOfColumn("name")
	angles := skinsData.GetIndexOfColumn("jet_angle")
	distances := skinsData.GetIndexOfColumn("jet_distance")
	body := skinsData.GetIndexOfColumn("body_artwork")
	forces := skinsData.GetIndexOfColumn("body_force_armor_artwork")
	drones := skinsData.GetIndexOfColumn("drone_artwork")
	credits := skinsData.GetIndexOfColumn("credit")
	customSkinCSVContents := skinsData.GetContents()

	discordUIDs := getDiscordUIDs()
	infoMaps := []map[string]string{discordUIDs}
	mapType := []cred.CreditSource{cred.Discord}

	skins = make([]CustomSkin, 0, skinsData.Rows())
	reqLength := skinsData.NumHeaders()

	for _, s := range customSkinCSVContents {
		if len(s) == reqLength {
			// print(i, v, body, forces, drones)

			name := s[names]
			distance := s[distances]
			angle := s[angles]
			skin := NewCustomSkin(name, distance, angle).addBody(s[body]).addForceA(s[forces]).addDrone(s[drones])

			credit, creditInfo, creditType := assignCredits(&s, credits, infoMaps, mapType)
			if !reflect.DeepEqual(creditType, "default") {
				skin.addCredits(cred.NewCredit(credit, creditInfo, creditType))
			}

			skins = append(skins, *skin)

			// printf("appropriate length: %d, %s", len(v), skin)
		} else {
			// printf("malformed csv, required length: %d, length: %d, %s,", reqLength, len(s), s)
		}
	}

	return
}

func assignCredits(s *[]string, cI int, maps []map[string]string, mapTypes []cred.CreditSource) (credit, creditInfo string, creditType cred.CreditSource) {
	// assign credits
	credit = (*s)[cI]

	for i, m := range maps {
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
	discordCreditData := fileIO.ReadCSV("DISCORD_UIDS")
	fileContents := discordCreditData.GetContents()

	uidMap := make(map[string]string, discordCreditData.Rows())

	for _, row := range fileContents {
		discordName := row[0]
		UID := row[1]
		uidMap[discordName] = UID
	}

	return uidMap
}

// ~~~~~~~~~~~~~~~~~~~ AssetPage

type AssetsPage struct {
	fileIO.MarkdownFile
	pageNumber int
	maxSkins   int
	skinsC     int

	skins []CustomSkin
}

func NewAssetsPage(filename string, pageNum int, path string) *AssetsPage {
	return &AssetsPage{MarkdownFile: *fileIO.NewMarkdownFile(filename, path), pageNumber: pageNum, maxSkins: 10, skinsC: 0}
}

func (a *AssetsPage) String() string {
	return a.GetFileName()
}

func (a *AssetsPage) bufferPagePreffix() error {
	// write to file:
	// Page #
	a.Append(fmt.Sprintf("# Page %d", a.pageNumber))
	// prev next
	err := a.bufferPrevNextPage()

	return err
}

func (a *AssetsPage) bufferPageSuffix() error {
	// write to file:
	// prev next
	err := a.bufferPrevNextPage()

	return err
}

func (a *AssetsPage) bufferPrevNextPage() error {
	path := "./"

	prev := format("Page_%d", a.pageNumber-1)
	prevF := format("%s.md", prev)
	curr := format("Page_%d", a.pageNumber)
	currF := format("%s.md", curr)
	next := format("Page_%d", a.pageNumber+1)
	nextF := format("%s.md", next)

	if a.pageNumber > 1 {

		a.AppendMarkdownLink(prev, (path + prevF))
	}

	a.AppendMarkdownLink(curr, (path + currF))
	a.AppendMarkdownLink(next, (path + nextF))

	return nil
}

func (a *AssetsPage) bufferCustomSkins(download bool) {
	// this writes to the custom skins stuff and adds the data, in markdown
	path := "https://github.com/areon546/NovaDriftCustomSkinRepository/raw/main"

	for _, skin := range a.skins {
		a.AppendNewLine()

		a.Append(format("**%s**: %s", skin.name, skin.FormatCredits()))
		a.AppendNewLine()

		a.Append("`" + skin.toCSVLine() + "`")
		a.AppendNewLine()

		a.AppendMarkdownEmbed(fileIO.ConstructPath(path, "custom_skins", skin.body))
		a.AppendMarkdownEmbed(fileIO.ConstructPath(path, "custom_skins", skin.forceArmour))
		a.AppendMarkdownEmbed(fileIO.ConstructPath(path, "custom_skins", skin.drone))
		// TODO append links to media  but how do we determine if there are media files?

		if download {
			a.AppendMarkdownLink("Download Me", fileIO.ConstructPath(path, "assets", format("%s.zip", skin.name)))
		}

		a.AppendNewLine()
	}
}

func (a *AssetsPage) writeBuffer() {
	a.WriteFile()

	// print(a.contentBuffer)
	helpers.Print("Writing to: ", a)
}

func (a *AssetsPage) addCustomSkins(cs []CustomSkin) {
	numSkins := min(10, len(cs))
	for a.skinsC < numSkins {
		a.skins = append(a.skins, cs[a.skinsC])
		a.skinsC++
	}
}

func ConstructAssetPages(skins []CustomSkin) (pages []AssetsPage) {
	numSkins := len(skins)
	// print("skins ", numSkins)
	numFiles := numSkins / 10

	if numSkins%10 != 0 {
		numFiles++
	}
	// print("filesToCreate", numFiles)

	for i := range numFiles {
		// create a new file
		pageNum := i + 1
		a := NewAssetsPage(fileIO.ConstructPath("", pagesFolder(), format("Page_%d", pageNum)), pageNum, "2")

		a.bufferPagePreffix()

		skinSlice, err := getNextSlice(skins, i)
		helpers.Handle(err)

		a.addCustomSkins(skinSlice)
		a.bufferCustomSkins(false)
		a.bufferPageSuffix()

		pages = append(pages, *a)
		// print(a)

		a.writeBuffer()
	}

	// a := NewAssetsPage(constructPath("", getPagesFolder(), "test"), 0, "")

	// a.bufferPagePreffix()
	// a.addCustomSkins(skins)
	// a.bufferCustomSkins()
	// a.bufferPageSuffix()

	// pages = append(pages, *a)
	return
}

func getNextSlice(skins []CustomSkin, i int) (subset []CustomSkin, err error) {
	numSkins := len(skins)

	if i < 0 || i > (len(skins)/10+1) {
		err = errors.New("index out of bounds for CustomSkins array")
	}

	min, max := i*10, (i+1)*10

	if max > numSkins {
		max = numSkins
	}

	return skins[min:max], err
}

func ConstructZipFiles(skins []CustomSkin) []fileIO.File {
	return make([]fileIO.File, 0)
}
