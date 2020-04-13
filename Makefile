init_repo: clean
	git init --bare test-repo.git
	echo "test" >> readme &
	git --work-tree=./test-repo.git --git-dir=./test-repo.git add .
	git --work-tree=./test-repo.git --git-dir=./test-repo.git commit -am "test file"
	echo "test" >> readme
	git --work-tree=./test-repo.git --git-dir=./test-repo.git add .
	git --work-tree=./test-repo.git --git-dir=./test-repo.git commit -am "test file"

clean:
	rm -rf ./test-repo.git

build:
	go build -mod=vendor

test: build init_repo
	./libgit2-rename-test