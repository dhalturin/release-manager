# Release Manager (slack > gitlab)

> Tool for manage releases in GitLab repository

## Prepare

- Create slack app on https://api.slack.com/apps
- Enable functionality
  - Interactive Components
  - Bots
  - Slash commands
- Edit scopes on permissons page
  - chat:write:bot
  - chat:write:user
  - incoming-webhook
  - bot
  - commands
  - usergroups:read
- Get OAuth Access Token
- Get Verification Token

## Build

For build package go in package directory and run `make`

```
make
if [ -z ************************ ]; then echo "Slack verification token is empty"; exit 1; fi
if [ -z xoxp-**********-************-************-******************************** ]; then echo "Slack OAuth token is empty"; exit 1; fi
if [ -z https://example.com ]; then echo "GitLab host is empty"; exit 1; fi
go build \
		-o usr/local/sbin/release-manager \
  		-ldflags "\
		  -X github.com/dhalturin/release-manager/data.SlackVerification=************************ \
		  -X github.com/dhalturin/release-manager/data.SlackOAuth=xoxp-**********-************-************-******************************** \
		  -X github.com/dhalturin/release-manager/server.GitLabHost=https://example.com"
```

## Command list

For showing command list you need use `/rm help` in slack channel

```
/rm job abort - abort running jobs from last pipeline of current release
/rm job list - run jobs from last pipeline of current release
/rm job run - run jobs from last pipeline of current release
/rm release add - create new release branch
/rm release del - removing release branch
/rm release edit - edit information of current release (only notes)
/rm release get - get last release
/rm release publish - publish release information to channel
/rm repo add - adding new repository
/rm repo choose - select the repository to perform actions
/rm repo del - removing repository
/rm repo edit - update existing repository
/rm task add - add task branch to release branch
```
