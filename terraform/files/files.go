package files

import (
	"sort"
	"strings"
)

const (
	TF          = ".tf"
	TFVARS      = ".tfvars"
	AUTO_TFVARS = ".auto.tfvars"

	DEFAULT_TFVARS = "terraform.tfvars"
)

var VALID_VARIABLE_FILES = [...]string{TF, AUTO_TFVARS}
var VALID_TERRAFORM_FILES = [...]string{TF}

func isTerraformTfvarsFile(fileName string) bool {
	// the terraform.tfvars file is a strict file name so make sure the file isn't called something like *terraform.tfvars
	return fileName == DEFAULT_TFVARS || strings.HasSuffix(fileName, "/"+DEFAULT_TFVARS)
}

func IsValidVariableFile(fileName string) bool {
	if isTerraformTfvarsFile(fileName) {
		return true
	}
	for _, fileExt := range VALID_VARIABLE_FILES {
		if strings.HasSuffix(fileName, fileExt) {
			return true
		}
	}
	return false
}

func IsValidTerraformFile(fileName string) bool {
	for _, fileExt := range VALID_TERRAFORM_FILES {
		if strings.HasSuffix(fileName, fileExt) {
			return true
		}
	}
	return false
}

type PrioritisableFile struct {
	fileName string
	priority int
}

func createPrioritisableFile(fileName string) PrioritisableFile {
	// TODO: Variables in terraform.tfvars.json files come after terraform.tfvars and before .auto.tfvars (not supported)
	// TODO: Variables in -var and -var-file options come after .auto.tfvars and .auto.tfvars.json files, in the order they are provided (not supported)

	// The file with the biggest value has the highest priority
	if strings.HasSuffix(fileName, TF) {
		// Default values have lowest priority (in .tf files)
		return PrioritisableFile{
			fileName: fileName,
			priority: 1,
		}
	} else if isTerraformTfvarsFile(fileName) {
		// Then variables in the terraform.tfvars file if it exists
		return PrioritisableFile{
			fileName: fileName,
			priority: 2,
		}
	} else if strings.HasSuffix(fileName, AUTO_TFVARS) {
		// Then variables in .auto.tfvars, in lexical order
		// TODO: or.auto.tfvars.json (not supported)
		return PrioritisableFile{
			fileName: fileName,
			priority: 3,
		}
	} else {
		// Won't happen
		return PrioritisableFile{
			fileName: fileName,
			priority: 0,
		}
	}
}

func OrderFilesByPriority(fileNames []string) []string {
	prioritisableFiles := make([]PrioritisableFile, 0, len(fileNames))

	for _, fileName := range fileNames {
		prioritisableFiles = append(prioritisableFiles, createPrioritisableFile(fileName))
	}

	sort.Slice(prioritisableFiles, func(i, j int) bool {
		// sort random files as the lowest priority
		if prioritisableFiles[i].priority == 0 || prioritisableFiles[j].priority == 0 {
			return i >= j
		}

		if prioritisableFiles[i].priority == prioritisableFiles[j].priority {
			// sort files with the same priority and .auto.tfvars lexically
			if strings.HasSuffix(prioritisableFiles[i].fileName, AUTO_TFVARS) {
				x := strings.Compare(prioritisableFiles[i].fileName, prioritisableFiles[j].fileName)
				return x <= 0
			}
			// do not sort files with the same priority and non .auto.tfvars
			return i <= j
		}

		return prioritisableFiles[i].priority < prioritisableFiles[j].priority
	})

	prioritisedFileNames := make([]string, 0, len(fileNames))
	for _, prioritisableFile := range prioritisableFiles {
		prioritisedFileNames = append(prioritisedFileNames, prioritisableFile.fileName)
	}

	return prioritisedFileNames
}