# K8s DevOps console

## Build (docker)

```bash
docker build -t webdevops/k8s-devops-console .
docker run --rm -p 9000:9000 -e KUBECONFIG=/root/.kube/config -v /path/to/your/kubeconfig:/root/.kube/config webdevops/k8s-devops-console
```

## Login (example)

Username: admin
Password: admin
-> Access to team `admin` and `user`

Username: user
Password: user
-> Access to team `user`

## TODO
- setup user on login
- setup user role on namespace creation
- create user key (store in remote storage?)
- cron user cleanup
- send notifcations
- search
