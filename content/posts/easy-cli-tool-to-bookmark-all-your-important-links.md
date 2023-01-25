---
title: "Easy Cli Tool to Bookmark All Your Important Links"
date: 2023-01-25T17:21:31+05:30
draft: false
---


### Create a cli tool to bookmark all your important stuff and never forget a link 

I have a additional blog [Intersting Posts]({{< ref "intersting_posts.md#reference" >}}) on my page where I keep all my important links for links that I feel are useful to me . But I usually do not have time to go through each of them.  As there is only so much useful blogs out there that it is very hard to keep track of the ones. 

Imagine if you could do this 
```
posts https://some-post post-name group-name 
```
in your cli and it will add that particular post to your reading list which is a markdown file  and also push the changes to a remote vcs like github. 

Today we will try to recreate something like this and we will use golang for this . 

```
mkdir posts
cd posts 
go mod init github.com/posts 
touch main.go
vi main.go 
```

Here I am using vim , but you are welcome to use any text-editor of your choice 

Ah I almost forgot . Create a Makefile as well , just to make your life easier


```
touch Makefile 
vi Makefile 
```

Inside make file  add a build command 
```
build:
	go build  -o posts  ./cli/main.go

```

Now we are ready to write the tool , but before that some helper functions and structs that would help us in our task 


{{< highlight go "linenos=inline,linenostart=1" >}}


type Group struct {
	Name  string
	links []string
}

// add the new link to the group 
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
// write the new data to the file
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
// function to parse content and combine them into groups 
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
// helper function to stage, commit and push our changes. So do a git init and setup a remote branch before doing so
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


{{< / highlight >}}

Now the main func 

{{< highlight go "linenos=inline,linenostart=1" >}}
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
	// Here I am reading all the previous posts that are in my reading list 
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
	// parse content of the reading list and combine them into groups 
	group_maps := parseContent(string(content))

	// add the new content to the reading list 
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
	// truncate the file as you want to overwrite them and write to the new file
	writeToFile(group_maps, file)
	file.Close()
	fmt.Println(commit_message)
	// do a git commit and push 
	commitAndPush(commit_message)
}

{{< / highlight >}}


Once done do a 
```
make build
```
to build your executable posts 

and run 
```
posts https://test-post <test-post> <group-name> <commit-message>
```
and it would add the corresponding post in the group-name 

With this you will never lose track of your bookmarks again 😃

We could have used [cobra](https://github.com/spf13/cobra) for parsing the commands but that would be a overkill . So decided against it . 


Do share this if you found it useful .Bye  