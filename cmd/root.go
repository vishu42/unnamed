/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// TreeNode represents a node in the tree
type TreeNode struct {
	Name     string
	Children []*TreeNode
}

type Workflow struct {
	Name      string
	Path      string
	Variables Variables
	Prefix    string
}

// NewTreeNode creates a new TreeNode
func NewTreeNode(name string) *TreeNode {
	return &TreeNode{
		Name:     name,
		Children: []*TreeNode{},
	}
}

// AddChild adds a child node to the current node
func (n *TreeNode) AddChild(child *TreeNode) {
	n.Children = append(n.Children, child)
}

// BuildDirectoryTree builds a tree from a directory path
func BuildDirectoryTree(dir string) (*TreeNode, error) {
	root := NewTreeNode(dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != dir {
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			currentNode := root
			for _, part := range filepath.SplitList(relPath) {
				childNode := findOrCreateChild(currentNode, part)
				currentNode = childNode
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return root, nil
}

// findOrCreateChild finds a child node by name or creates a new one if it doesn't exist
func findOrCreateChild(parent *TreeNode, name string) *TreeNode {
	for _, child := range parent.Children {
		if child.Name == name {
			return child
		}
	}

	newChild := NewTreeNode(name)
	parent.AddChild(newChild)
	return newChild
}

// PrintTree recursively prints the tree structure
func PrintTree(node *TreeNode, indent string) {
	fmt.Println(indent + node.Name)
	for _, child := range node.Children {
		PrintTree(child, indent+"  ")
	}
}

//go:embed sample-workflow.yaml
var tmplWorkflow string
var (
	TemplateDir string
	rootCmd     = &cobra.Command{
		Use:   "unnamed",
		Short: "Generates github workflows corresponding to each terraform template",
		// Uncomment the following line if your bare application
		// has an action associated with it:
		Run: func(cmd *cobra.Command, args []string) {
			w := ParseTemplateDirectory(TemplateDir)
			GenerateWorkflow(&w)
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

/*

- parse root directory and store data in a tree data structure
- walk over each direct path from root node to leaf node
- generate a workflow based on variable.tf file present at the leaf directory

*/

func GenerateWorkflow(w *Workflow) {
	tmpl := template.New("workflow").Delims("[[", "]]")

	// Add a custom function to the template for converting text to uppercase
	tmpl = tmpl.Funcs(template.FuncMap{
		"toUpper": strings.ToUpper,
	})

	// Parse the template content
	tmpl, err := tmpl.Parse(tmplWorkflow)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Render the template to standard output
	err = tmpl.Execute(os.Stdout, &w)
	if err != nil {
		log.Fatal("Error executing template:", err)
	}
}

// findRepoRoot walks up the directory tree until it finds the .git directory
func findRepoRoot(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); !os.IsNotExist(err) {
			return dir, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", fmt.Errorf("no .git directory found")
		}
		dir = parentDir
	}
}

func ParseTemplateDirectory(dir string) (w Workflow) {
	if !filepath.IsAbs(dir) {
		ap, err := filepath.Abs(dir)
		if err != nil {
			fmt.Println(err)
			return
		}

		dir = ap
	}

	// logic for each terminal directory in the directory tree

	v, err := ParseVariables(dir)
	if err != nil {
		panic(err)
	}

	// find repo root
	rootPath, err := findRepoRoot(dir)
	if err != nil {
		panic(err)
	}

	relPath, err := filepath.Rel(rootPath, dir)
	if err != nil {
		panic(err)
	}

	w.Name = relPath
	w.Path = relPath
	w.Variables = v

	fmt.Println(relPath)
	w.Prefix = strings.ReplaceAll(relPath, "/", "_") + "_"

	fmt.Printf("%+v", w)

	return
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVarP(&TemplateDir, "template-directory", "r", ".", "root directory for templates")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
