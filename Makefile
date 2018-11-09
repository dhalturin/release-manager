all: build

SlackVerification=************************
SlackOAuth=xoxp-**********-************-************-********************************
GitLabHost=https://example.com

clean:
	rm -f ./usr/local/sbin/*

build:
	if [ -z $(SlackVerification) ]; then echo "Slack verification token is empty"; exit 1; fi
	if [ -z $(SlackOAuth) ]; then echo "Slack OAuth token is empty"; exit 1; fi
	if [ -z $(GitLabHost) ]; then echo "GitLab host is empty"; exit 1; fi

	go build \
		-o usr/local/sbin/release-manager \
  		-ldflags "\
		  -X github.com/dhalturin/release-manager/data.SlackVerification=$(SlackVerification) \
		  -X github.com/dhalturin/release-manager/data.SlackOAuth=$(SlackOAuth) \
		  -X github.com/dhalturin/release-manager/data.Version=0.1-dev \
		  -X github.com/dhalturin/release-manager/server.GitLabHost=$(GitLabHost)"

run: build
	./usr/local/sbin/release-manager


go run -ldflags "-X github.com/dhalturin/release-manager/data.SlackVerification=$(SlackVerification) -X github.com/dhalturin/release-manager/data.SlackOAuth=$(SlackOAuth) -X github.com/dhalturin/release-manager/server.GitLabHost=$(GitLabHost)"
