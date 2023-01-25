package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// this should be a cli tool for getting two things
	// posts https://some-link link-name
	// do a git add . commit and push to repo and the ci/cd will do its things for you

	if len(os.Args) < 2 {
		fmt.Println("Usage: posts https://some-link link-name group-name commit-message")
		os.Exit(1)
	}

	group_name := "Default"
	commit_message := "Added new link"
	if len(os.Args) >= 4 {
		group_name = os.Args[3]
	}

	if len(os.Args) >= 5 {
		commit_message = os.Args[4]
	}

	added_link := fmt.Sprintf("- [%s](%s)", os.Args[2], os.Args[1])

	file, err := os.OpenFile("./content/posts/intersting_posts.md", os.O_RDONLY, os.ModePerm)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	group_maps := parseContent(string(content))
	addContent(group_maps, group_name, added_link)
	err = file.Close()
	if err != nil {
		fmt.Println(err)
	}
	file, err = os.OpenFile("./content/posts/intersting_posts.md", os.O_WRONLY|os.O_TRUNC, os.ModePerm)

	if err != nil {
		fmt.Println(err)

	}
	file.Truncate(0)
	writeToFile(group_maps, file)
	file.Close()
	fmt.Println(commit_message)
	commitAndPush(commit_message)
}

func addContent(group_maps map[string]*Group, group_name, added_link string) {
	group_name = "### " + group_name
	curr_group, ok := group_maps[group_name]
	if ok {
		curr_group.links = append(curr_group.links, added_link)
	} else {
		curr_group = &Group{
			Name:  group_name,
			links: []string{added_link},
		}
		group_maps[group_name] = curr_group
	}
}

func writeToFile(group_maps map[string]*Group, file *os.File) {
	writer := bufio.NewWriter(file)
	writer.WriteString("---\n")
	writer.WriteString("title : 'Interesting posts'\n")
	writer.WriteString("date: 2023-01-25T14:42:15+05:30\n")
	writer.WriteString("draft: false\n")
	writer.WriteString("---\n")
	writer.WriteString("\n\n")
	writer.WriteString("## List of interesting blog posts and videos\n")
	writer.WriteString("__________\n")
	for group_name, group := range group_maps {
		writer.WriteString(group_name + "\n")
		for _, link := range group.links {
			writer.WriteString(link + "\n")
		}
		writer.WriteString("\n\n")
	}
	err := writer.Flush()
	if err != nil {
		fmt.Println("Cannot write to file")
		fmt.Println(err)
		return
	}
}

func parseContent(content string) map[string]*Group {
	lines := strings.Split(content, "\n")
	group_maps := map[string]*Group{}
	curr_group := "Default"
	for _, line := range lines {
		if strings.HasPrefix(line, "###") {
			curr_group = line
		} else if strings.HasPrefix(line, "- [") {
			group, ok := group_maps[curr_group]
			if ok {
				group.links = append(group.links, line)
			} else {
				group = &Group{
					Name:  curr_group,
					links: []string{line},
				}
				group_maps[curr_group] = group
			}
		}
	}
	return group_maps
}

func commitAndPush(commit_message string) {

	add_cmd := exec.Command("git", "add", ".")
	add_cmd.Start()
	if err := add_cmd.Wait(); err != nil {
		fmt.Println(err)
	}

	commit_cmd := exec.Command("git", "commit", "-m", fmt.Sprintf("'%s'", commit_message))
	commit_cmd.Start()

	if err := commit_cmd.Wait(); err != nil {
		fmt.Println(err)
	}

	push_cmd := exec.Command("git", "push", "origin")
	push_cmd.Start()
	if err := push_cmd.Wait(); err != nil {
		fmt.Println(err)
	}

}

type Group struct {
	Name  string
	links []string
}
