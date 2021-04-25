#======需要设置=======
export dr_user=?
export dr_password=?
#===================

echo "添加执行权限"
chmod u+x /tmp/app/inetd

echo "后台运行"
nohup /tmp/app/inetd 1 >/dev/null 2>&1 </dev/null &

echo "删除日志文件"
rm -rf nohup.out
