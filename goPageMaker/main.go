package main

import (
	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/fileIO"
	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/helpers"
	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/nova"
)

func main() {
	testing := false

	if !testing {
		print("Testing")

		test()

		return
	}

	print("Running")

	// delete the entirety of the pages' folder's contents if present

	// returns a list of CustomSkins based on whats in the custom_skins folder
	print("Compiling Skins")
	skins := nova.GetCustomSkins()

	// print(skins)
	_ = nova.ConstructZipFiles(skins) // zipFiles

	nova.ConstructAssetPages(skins)

}

func test() {
	fileIO.ZipFolder("../goPageMaker/helpers", "../helpers")
}

func print(a ...any) {
	helpers.Print(a)
}
