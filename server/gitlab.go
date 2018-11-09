package server

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"

	gitlab "github.com/xanzy/go-gitlab"
)

// BranchList struct
type BranchList struct {
	Name string
	ID   int
}

func gitLabListBranches(c *gitlab.Client, pid interface{}, branches *[]BranchList, pattern string) error {
	process := true
	page := 1

	for process {
		log.Print("get page: ", page)

		opts := gitlab.ListBranchesOptions{
			Page:    page,
			PerPage: 100,
		}

		branch, response, err := c.Branches.ListBranches(pid, &opts)
		if err != nil {
			return err
		}

		for _, j := range branch {
			branchID := 0

			if pattern != "" {
				reg := regexp.MustCompile(fmt.Sprintf("^%s", pattern))

				if !reg.MatchString(j.Name) {
					continue
				}

				if id, err := strconv.Atoi(reg.FindStringSubmatch(j.Name)[1]); err == nil {
					branchID = id
				} else {
					return err
				}
			}

			*branches = append(*branches, BranchList{j.Name, branchID})
		}

		if response.Header.Get("X-Next-Page") == "" {
			process = false
		} else {
			page++
		}
	}

	sort.SliceStable(*branches, func(i, j int) bool {
		branch := *branches
		return branch[i].ID >= branch[j].ID
	})

	return nil
}
