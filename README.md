# 🏊‍♂️ Swim
## 🚀 Motivation
Did you ever want to change the port of your running Docker container? Well, now you can!

## 📦 Installation
```shell
go install github.com/4thel00z/swim/...@latest
```

## 🛠 Usage

```shell
# Give the ports you want to add
# You will be interactively prompted for the container to change
swim update-port -p hostIP:hostPort:containerPort -p hostIP:hostPort:containerPort
```

To provide a container ID directly and skip the interactive list:

```shell
swim update-port <containerID> -p hostIP:hostPort:containerPort -p hostIP:hostPort:containerPort
```

## 📜 License

This project is licensed under the GPL-3 license.