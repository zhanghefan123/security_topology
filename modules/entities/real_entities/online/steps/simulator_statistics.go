package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"zhanghefan123/security_topology/utils/dir"
	"zhanghefan123/security_topology/utils/file"
)

type DescriptionAndData struct {
	Description string
	Datas       string
}

// GetStatistics get simulator statistics after the end of simulation
func (s *Simulator) GetStatistics(destinationDir string) error {
	// judge if the destination directory exists, if not, create it
	if !dir.IsDirExists(destinationDir) {
		err := os.MkdirAll(destinationDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create dir failed due to: %v", err)
		}
	}

	err := s.GetPvLinksLegalRatios(destinationDir)
	if err != nil {
		return err
	}
	err = s.GetPvLinksGains(destinationDir)
	if err != nil {
		return err
	}

	err = s.GetPvLinksWeights(destinationDir)
	if err != nil {
		return err
	}

	err = s.GetPvLinksExploreProbabilities(destinationDir)
	if err != nil {
		return err
	}

	err = s.GetPathGains(destinationDir)
	if err != nil {
		return err
	}

	err = s.GetPathWeights(destinationDir)
	if err != nil {
		return err
	}

	err = s.GetPathExploreProbabilities(destinationDir)
	if err != nil {
		return err
	}

	err = s.GetSelectedPath(destinationDir)
	if err != nil {
		return err
	}

	err = s.GetRegrets(destinationDir)
	if err != nil {
		return err
	}
	return nil
}

func (s *Simulator) formatFloatSlice(values []float64) string {
	result := ""
	for index, value := range values {
		if index != len(values)-1 {
			result += fmt.Sprintf("%f,", value)
		} else {
			result += fmt.Sprintf("%f", value)
		}
	}
	return result
}

func (s *Simulator) writeStatisticsToFile(destinationDir, filename string, descAndDatas []DescriptionAndData) error {
	finalString := ""
	for _, descAndData := range descAndDatas {
		finalString += fmt.Sprintf("%s:%s\n", descAndData.Description, descAndData.Datas)
	}
	err := file.WriteStringIntoFile(filepath.Join(destinationDir, filename), finalString)
	if err != nil {
		return fmt.Errorf("failed to write %s into file: %v", filename, err)
	}
	return nil
}

// GetPvLinksLegalRatios 获取每条 directed pv link 在每个 epoch 的 legal ratio
func (s *Simulator) GetPvLinksLegalRatios(destinationDir string) error {
	resultList := make([]DescriptionAndData, 0)
	for _, absLink := range s.SimGraph.SimDirectedAbsLinks {
		descAndData := DescriptionAndData{
			Description: absLink.Description,
			Datas:       s.formatFloatSlice(absLink.LegalRatios),
		}
		resultList = append(resultList, descAndData)
	}
	return s.writeStatisticsToFile(destinationDir, "pv_links_legal_ratio.txt", resultList)
}

// GetPvLinksGains 获取每条 directed pv link 在每个 epoch 的 gains
func (s *Simulator) GetPvLinksGains(destinationDir string) error {
	resultList := make([]DescriptionAndData, 0)
	for _, pvLink := range s.SimGraph.SimDirectedAbsLinks {
		descAndData := DescriptionAndData{
			Description: pvLink.Description,
			Datas:       s.formatFloatSlice(pvLink.RectifiedGains),
		}
		resultList = append(resultList, descAndData)
	}
	return s.writeStatisticsToFile(destinationDir, "pv_links_gains.txt", resultList)
}

// GetPvLinksWeights 获取每条 directed pv link 在每个 epoch 的 weights
func (s *Simulator) GetPvLinksWeights(destinationDir string) error {
	resultList := make([]DescriptionAndData, 0)
	for _, pvLink := range s.SimGraph.SimDirectedAbsLinks {
		descAndData := DescriptionAndData{
			Description: pvLink.Description,
			Datas:       s.formatFloatSlice(pvLink.Weights),
		}
		resultList = append(resultList, descAndData)
	}
	return s.writeStatisticsToFile(destinationDir, "pv_links_weights.txt", resultList)
}

// GetPvLinksExploreProbabilities 获取每条 directed pv link 在每个 epoch 的 explore probability
func (s *Simulator) GetPvLinksExploreProbabilities(destinationDir string) error {
	resultList := make([]DescriptionAndData, 0)
	for _, pvLink := range s.SimGraph.SimDirectedAbsLinks {
		descAndData := DescriptionAndData{
			Description: pvLink.Description,
			Datas:       s.formatFloatSlice(pvLink.ExploreProbabilities),
		}
		resultList = append(resultList, descAndData)
	}
	return s.writeStatisticsToFile(destinationDir, "pv_links_explore_probabilities.txt", resultList)
}

// GetPathGains 获取每条路径在每个 epoch 的 gains
func (s *Simulator) GetPathGains(destinationDir string) error {
	resultList := make([]DescriptionAndData, 0)
	for _, simPath := range s.SimGraph.AvailablePaths {
		descAndData := DescriptionAndData{
			Description: simPath.Description,
			Datas:       s.formatFloatSlice(simPath.Gains),
		}
		resultList = append(resultList, descAndData)
	}
	return s.writeStatisticsToFile(destinationDir, "path_gains.txt", resultList)
}

// GetPathWeights 获取每条路径在每个 epoch 的 weights
func (s *Simulator) GetPathWeights(destinationDir string) error {
	resultList := make([]DescriptionAndData, 0)
	for _, simPath := range s.SimGraph.AvailablePaths {
		descAndData := DescriptionAndData{
			Description: simPath.Description,
			Datas:       s.formatFloatSlice(simPath.Weights),
		}
		resultList = append(resultList, descAndData)
	}
	return s.writeStatisticsToFile(destinationDir, "path_weights.txt", resultList)
}

// GetPathExploreProbabilities 获取每条路径在每个 epoch 的 explore probability
func (s *Simulator) GetPathExploreProbabilities(destinationDir string) error {
	resultList := make([]DescriptionAndData, 0)
	for _, path := range s.SimGraph.AvailablePaths {
		descAndData := DescriptionAndData{
			Description: path.Description,
			Datas:       s.formatFloatSlice(path.ExploreProbabilities),
		}
		resultList = append(resultList, descAndData)
	}
	return s.writeStatisticsToFile(destinationDir, "path_explore_probabilities.txt", resultList)
}

// GetSelectedPath 获取每个 epoch 选择的路径的 id
func (s *Simulator) GetSelectedPath(destinationDir string) error {
	resultList := make([]DescriptionAndData, 0)
	finalString := ""
	for index, simPath := range s.SimGraph.SelectedPaths {
		if index != len(s.SimGraph.SelectedPaths)-1 {
			finalString += fmt.Sprintf("%d,", simPath.PathId)
		} else {
			finalString += fmt.Sprintf("%d", simPath.PathId)
		}
	}
	descAndData := DescriptionAndData{
		Description: "selected path",
		Datas:       finalString,
	}
	resultList = append(resultList, descAndData)
	return s.writeStatisticsToFile(destinationDir, "selected_paths.txt", resultList)
}

func (s *Simulator) GetRegrets(destinationDir string) error {
	resultList := make([]DescriptionAndData, 0)
	descAndData := DescriptionAndData{
		Description: "regrets",
		Datas:       s.formatFloatSlice(s.SimGraph.Regrets),
	}
	resultList = append(resultList, descAndData)
	return s.writeStatisticsToFile(destinationDir, "regrets.txt", resultList)
}
