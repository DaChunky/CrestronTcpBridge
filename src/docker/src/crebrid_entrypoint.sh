echo create config
mkdir /etc/crebrid
touch /etc/crebrid/crebrid.conf
chmod 666 /etc/crebrid/crebrid.conf
parIP=${1:-localhost} 
parPort=${2:-65432} 
parAccessCode=${3:-} 
echo "# crebrid configuration file" \
    $'\n'ip=$parIP \
    $'\n'port=$parPort \
    $'\n'accessCode=$parAccessCode > /etc/crebrid/crebrid.conf
echo create log locations
mkdir /var/log/crebrid
chmod 744 /var/log/crebrid
touch /var/log/crebrid/crebrid.log
chmod 744 /var/log/crebrid/crebrid.log
cd /home/
./crebrid
#cat /var/log/crebrid/crebrid.log
