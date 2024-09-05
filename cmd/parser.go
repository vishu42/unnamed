package cmd

import (
	"fmt"
	"os"
	"path"

	hcl "github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Variables struct {
	Variables          []string
	SensitiveVariables []string
}

func ParseVariables(dir string) (v Variables, err error) {
	// Get a list of files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// init new hcl parser
	p := hcl.NewParser()
	// Iterate over the files and print their names
	for _, file := range files {
		// Check if the file is a regular file (not a directory)
		if !file.IsDir() {
			fmt.Println(file.Name())
			hcl, hcldiag := p.ParseHCLFile(path.Join(dir, file.Name()))
			if hcldiag.HasErrors() {
				fmt.Printf("Error parsing HCL file: %s\n", hcldiag)
				return
			}

			// Check if it's HCL syntax tree
			if body, ok := hcl.Body.(*hclsyntax.Body); ok {
				// Iterate over blocks in the file
				for _, block := range body.Blocks {
					if block.Type == "variable" {
						// fmt.Printf("Variable Name: %s\n", block.Labels[0])
						sensitive := false
						// check if var is sensitive and no default attribute is present
						att, ok := block.Body.Attributes["sensitive"]
						if ok {
							val, diags := att.Expr.Value(nil)
							if diags.HasErrors() {
								fmt.Printf("Error parsing attribute %s: %s\n", att, diags)
								continue
							}
							if val.True() {
								sensitive = true
							}
						}

						defaultPresent := false
						att, ok = block.Body.Attributes["default"]
						if ok {
							val, diags := att.Expr.Value(nil)
							if diags.HasErrors() {
								fmt.Printf("Error parsing attribute %s: %s\n", att, diags)
								continue
							}
							if val.AsString() != "" {
								defaultPresent = true
							}
						}

						if !defaultPresent && !sensitive {
							v.Variables = append(v.Variables, block.Labels[0])
						}

						if !defaultPresent && sensitive {
							v.SensitiveVariables = append(v.SensitiveVariables, block.Labels[0])
						}

					}
				}
			} else {
				fmt.Println("The file is not in the expected HCL format.")
			}
		}
	}
	return
}
