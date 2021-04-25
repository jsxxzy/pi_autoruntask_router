该项目意义不大, 只是用来在树莓派下自动登录(对抗不可抗力, 比如说断电), 具体流程是用`go`写一个`web`服务, 当路由器断电之后, 路由器定时任务将停止, 且会断网, 此时,通过树莓派执行`ssh`来判断进程是否存在, 不存在就将`inetd`和`daemon.sh`上传到路由器上去

在编译之前`daemon`设置为正确的哆点账号密码

```sh
#======需要设置=======
export dr_user=?
export dr_password=?
#===================
```

并且, 你需要重写 `func uploadDaemon()` 的上传逻辑

而且你需要自己手动编译 `daemon` 程序, 它在: https://github.com/jsxxzy/inet/cmd/daemon/router

在执行之前也需要将网管设置一下: `build.sh`

```bash
# 树莓派的局域网ip
export pi_ipv4=10.32.0.188
```

直接执行就完事了

```
chmod u+x ./build.sh
sg ./build.sh
```

`windows`需要`ssh` + `scp` 支持, 推荐使用 `git for windows bash`