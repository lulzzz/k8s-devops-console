# K8s DevOps console

## Build (docker)

```bash
docker build -t webdevops/k8s-devops-console .
docker run --rm -p 9000:9000 -e KUBECONFIG=/root/.kube/config -v /path/to/your/kubeconfig:/root/.kube/config webdevops/k8s-devops-console
```

## Environment settings

| Env var               |Required   | Type     | Description                                           |
|:----------------------|:----------|:---------|:------------------------------------------------------|
| KUBECONFIG            | no        | string   | Path to custom kubeconf (if not in-cluster)           |
| OAUTH_PROVIDER        | yes       | string   | OAuth provider name                                   |
| OAUTH_CLIENT_ID       | yes       | string   | OAuth client id                                       |
| OAUTH_CLIENT_SECRET   | yes       | string   | OAuth client secret                                   |

### OAuth

Supported providers:

- github ([create new application](https://github.com/settings/developers))

## TODO
- setup k8s permissions on login
- setup user role on namespace creation
- create user key (store in remote storage?)
- cron user cleanup
- send notifcations
