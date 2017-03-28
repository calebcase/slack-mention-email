# Slack Mention Email

A simplistic Slack client that watches for mentions and sends an email to you.

## Getting Slack Token

https://api.slack.com/custom-integrations/legacy-tokens

```
export SLACK_TOKEN=<token>
```

## Gmail SMTP

Goto your account settings and create an application login for `Mail`. This will provide you with an application specific password which you can use.

```
export SMTP_USER=<username>
export SMTP_PASS=<password>
export SMTP_DOMAIN=smtp.gmail.com
```

## Usage

```
go get github.com/calebcase/slack-mention-email
# Export the variables as above.
slack-mention-email
```

Now when an mention is detected you will receive an email.
