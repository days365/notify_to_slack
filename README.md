# Notify to Slack

Notify message to slack from gcs storage file.

## Deploy

```
$ gcloud functions deploy notify_to_slack \
    --entry-point NotifyToSlack \
    --runtime go121 \
    --set-env-vars 'API_TOKEN=...,CHANNEL=...' \
    --trigger-bucket <bucket> \
    --project <your_gcp_project_id> \
    --region asia-northeast1
```
