package main

import (
    "fmt"
    "log"
    "sync"

    "github.com/google/uuid"
    git "github.com/libgit2/git2go/v30"
)

const (
    TestRepoPath = "./test-repo.git"
    TestDefaultBranch = "master"
)

func main() {
    repo, err := git.OpenRepository(TestRepoPath)
    if err != nil {
        fmt.Println(err)
        return
    }

    branch, err := repo.LookupBranch(TestDefaultBranch, git.BranchLocal)
    if err != nil {
        log.Fatalln("lookup default branch failed")
    }
    defer branch.Free()

    commitsData := make([]string, 0, 2)
    commit, err := repo.LookupCommit(branch.Target())
    if err != nil {
        log.Fatalln("got default branch commit error")
    }
    defer commit.Free()

    ourCommit := commit.Id().String()
    commitsData = append(commitsData, commit.Id().String())
    commitsData = append(commitsData, commit.ParentId(0).String())

    wd := sync.WaitGroup{}
    wd.Add(2)

    go func() {
        defer wd.Done()
        for {
            iter, err := repo.NewBranchIterator(git.BranchLocal)
            if err != nil {
                fmt.Printf("Rename:NewBranchIterator error with %v", err)
                return
            }

            iter.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
                branchName, err := branch.Name()
                if err != nil {
                    fmt.Println(err)
                    return err
                }
                fmt.Println(branchName)
                uid := uuid.New()
                _, err = branch.Rename(fmt.Sprintf("refs/heads/%s", uid.String()), true, fmt.Sprintf("renamed from %s", branchName))
                if err != nil {
                    fmt.Printf("reaname error with %v",err)
                    return err
                }
                return nil
            })
        }

    }()

    go func(ourCommit string) {
        defer wd.Done()
        for {
           iter, err := repo.NewBranchIterator(git.BranchLocal)
           if err != nil {
               log.Printf("Rename:NewBranchIterator error with %v", err)
               return
           }

           newCommit := commitsData[0]

           err = iter.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
               thisCommit := branch.Target().String()
               ourCommit = thisCommit
               if thisCommit == newCommit {
                   newCommit = commitsData[1]
               }
               obj, err := repo.RevparseSingle(newCommit)
               if err != nil {
                   log.Printf("rev commit error with %v\n", err)
                   return err
               }
               _, err = branch.SetTarget(obj.Id(), fmt.Sprintf("set from %s", thisCommit))
               if err != nil {
                   log.Printf("set commit error with %v\n", err)
                   return err
               }
               ourCommit = obj.Id().String()
               return err
           })
           iter.Free()

           if err != nil {
               log.Printf("update commit error with %v\n", err)
               continue
           }

           iter2, err := repo.NewBranchIterator(git.BranchLocal)
           if err != nil {
               log.Printf("Rename:NewBranchIterator error with %v", err)
               return
           }

           iter2.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
              thisCommit := branch.Target().String()
              if thisCommit != ourCommit {
                  log.Fatalf("not equal, except %s got %s", ourCommit, thisCommit)
              }
              return nil
           })
           iter2.Free()
       }
    }(ourCommit)
    wd.Wait()
}
