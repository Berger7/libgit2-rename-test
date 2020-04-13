init_repo: clean
	git init --bare test-repo.git
	git --work-tree=./test-repo.git --git-dir=./test-repo.git commit --allow-empty -m "test1"
	git --work-tree=./test-repo.git --git-dir=./test-repo.git commit --allow-empty -m "test2"

clean:
	rm -rf ./test-repo.git

build:
	go build -mod=vendor

test: build init_repo
	./libgit2-rename-test