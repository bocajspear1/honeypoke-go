#!/bin/sh

if [ ! -f config.json ]; then
    echo "config.json must be created first"
    exit 1
fi 

echo "Getting GeoLite Database..."
wget https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz --quiet
tar -zxf GeoLite2-City.tar.gz
mv GeoLite2-City_*/GeoLite2-City.mmdb ./
rm GeoLite2-City.tar.gz
rm -rf GeoLite2-City_*/
echo ""
echo "Setting up 'large' directory..."
if [ -f config.json ]; then
    USER=$(grep '"user":' config.json | cut -d":" -f 2 | sed 's_[", ]__g')
    GROUP=$(grep '"group":' config.json | cut -d":" -f 2 | sed 's_[", ]__g')
    mkdir -p ./large
    sudo chown ${USER}:${GROUP} ./large
else
    echo "config.json does not exist, cannot create 'large' directory"
fi

echo ""
echo "Creating SSL certificate..."
openssl req -new -x509 -days 365 -nodes -out honeypoke_cert.pem -keyout honeypoke_key.pem