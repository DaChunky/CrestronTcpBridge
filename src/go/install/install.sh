echo copy executable to /bin/. ...
mv crebrid /bin/.
chmod 755 /bin/crebrid
mv crebri /bin/.
chmod 755 /bin/crebri
echo create config
mkdir /etc/crebrid
touch /etc/crebrid/crebrid.conf
chmod 666 /etc/crebrid/crebrid.conf
parIP="${1:-localhost}" 
parPort="${2:-65432}" 
parAccessCode="${3:-}" 
echo "# crebrid configuration file" \
    "\n#ip=$parIP" \
    "\n#port=$parPort" \
    "\n#accessCode=$parAccessCode" > /etc/crebrid/crebrid.conf
echo create log locations
mkdir /var/log/crebrid
chmod 744 /var/log/crebrid
touch /var/log/crebrid/crebrid.log
chmod 744 /var/log/crebrid/crebrid.log
mkdir /var/log/crebri
touch /var/log/crebri/crebri.log
chmod 777 /var/log/crebri
chmod 744 /var/log/crebri/crebri.log
echo copy service definition file...
mv crebrid.service /lib/systemd/system/.
chmod 755 /lib/systemd/system/crebrid.service

echo setup service...
systemctl daemon-reload
systemctl start crebrid.service
systemctl enable crebrid.service