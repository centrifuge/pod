package utils

func GetCentLogFormat() string {
	return `%{time:02.01.2006 15:04:05.000}  %{color:bold} %{level} %{color:reset} %{color:blue} %{module}: %{color:reset} %{message} %{shortfile}`

}
