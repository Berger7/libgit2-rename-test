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
        log.Fatalf("open repository failed with %v\n", err)
        return
    }

    branch, err := repo.LookupBranch(TestDefaultBranch, git.BranchLocal)
    if err != nil {
        log.Fatalf("lookup default branch failed with %v\n", err)
    }
    defer branch.Free()

    commitsData := make([]string, 0, 2)
    commit, err := repo.LookupCommit(branch.Target())
    if err != nil {
        log.Fatalf("got default branch commit failed with %v\n", err)
    }
    defer commit.Free()

    /*
     * The ourCommit represents the value of the branch in the warehouse in normal logic.
     * The ourCommit will only be modified in the logic to reset the commitID of the branch.
     * And after resetting successfully, assign ourCommit to the modified value.
     */
    ourCommit := commit.Id().String()

    commitsData = append(commitsData, commit.Id().String())
    commitsData = append(commitsData, commit.ParentId(0).String())

    // set two coroutine
    wg := sync.WaitGroup{}
    wg.Add(2)

    // rename an exist branch name to a random string.
    go func() {
        defer wg.Done()
        for {
            iter, err := repo.NewBranchIterator(git.BranchLocal)
            if err != nil {
                log.Printf("NewBranchIterator failed with %v", err)
                return
            }

            iter.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
                branchName, err := branch.Name()
                if err != nil {
                    log.Printf("get branch name failed with %v\n", err)
                    return err
                }

                // Use uuid as the random name of the branch.
                // The format is similar to refs/heads/e09ab860-5636-407c-bc46-95f2131fb59d
                uid := uuid.New()
                _, err = branch.Rename(fmt.Sprintf("refs/heads/%s", uid.String()), true,
                    fmt.Sprintf("renamed from %s", branchName))
                if err != nil {
                    log.Printf("reaname failed with %v",err)
                    return err
                }
                return nil
            })
        }

    }()

    /*
     *    This coroutine only does two things:
     *      1. Get the current branch of the repository and reset it's commitID. Since there is no operation to create a
     *         branch, there should only be one branch.
     *      2. If the first step is successful, reacquire the commitID of the current repository branch and compare it
     *         with the value set in the first step. Only this process will modify the commitID of the branch.Therefore,
     *         this value should be unchanged. Unless there are some problems in the renaming.
     */
    go func(ourCommit string) {
        defer wg.Done()
        for {
           iter, err := repo.NewBranchIterator(git.BranchLocal)
           if err != nil {
               log.Printf("Set:NewBranchIterator error with %v", err)
               return
           }

           // There will be no new commits in the repository, so just take a commit as the initial value
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

           // If the setting fails, skip without comparing
           if err != nil {
               log.Printf("update commit error with %v\n", err)
               continue
           }

           // Re-read the branch information of the repository
           iter2, err := repo.NewBranchIterator(git.BranchLocal)
           if err != nil {
               log.Printf("Check:NewBranchIterator failed with %v", err)
               return
           }

           // Fatal error will only be returned if the branch information is obtained and is inconsistent with the
           // set result.
           iter2.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
              thisCommit := branch.Target().String()
              if thisCommit != ourCommit {
                  log.Fatalf("not equal, expect %s got %s", ourCommit, thisCommit)
              }
              return nil
           })
           iter2.Free()
       }
    }(ourCommit)
    wg.Wait()
}
