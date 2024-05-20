# swim

## Motivation

Did you ever wanted to change the port of your *running* docker container?
Well now you can!

## Installation

```shell
go install github.com/4thel00z/swim/...@latest
```

## Usage

```shell
# Give the ports you want to add
# You will be interactively prompted for the container to change
swim update-port -p 9090:8080 -p 8081:8081
```

## License

This project is licensed under the GPL-3 license.
