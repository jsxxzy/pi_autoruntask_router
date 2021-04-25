
# 树莓派的局域网ip
export pi_ipv4=10.32.0.188

rm -rf $PWD/router
CGO_ENABLED=0 GOOS=linux GOARCH=arm go build .
scp $PWD/router pi@$pi_ipv4:/home/pi/data