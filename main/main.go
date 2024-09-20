package main

import (
	"os"
	"zhanghefan123/security_topology/main/cmd"
	"zhanghefan123/security_topology/main/cmd/constellation"
	"zhanghefan123/security_topology/main/cmd/images"
)

func main() {
	rootCmd := cmd.CreateRootCmd()
	constellationCmd := constellation.CreateConstellationCmd()
	imagesCmd := images.CreateImagesCmd()
	rootCmd.AddCommand(constellationCmd)
	rootCmd.AddCommand(imagesCmd)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
