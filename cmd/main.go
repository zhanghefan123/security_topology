package main

import (
	"os"
	"zhanghefan123/security_topology/cmd/constellation"
	"zhanghefan123/security_topology/cmd/images"
	"zhanghefan123/security_topology/cmd/root"
)

func main() {
	rootCmd := root.CreateRootCmd()
	constellationCmd := constellation.CreateConstellationCmd()
	imagesCmd := images.CreateImagesCmd()
	rootCmd.AddCommand(constellationCmd)
	rootCmd.AddCommand(imagesCmd)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
