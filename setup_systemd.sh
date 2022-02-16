#!/usr/bin/env bash
cur=$(dirname "$(readlink -f $0)")
# set -x
set -e

var_cur=$(dirname "$(readlink -f $0)")
var_service_filename=rinetd.service
var_dst_path=${var_cur}/${var_service_filename}


# Restart=always
# ExecStart=/bin/python xxx.py
# Type=simple
# Type=oneshot
# ExecStart=/usr/bin/sleep infinity

cat <<EOF > ${var_dst_path}
# copy to /usr/lib/systemd/system
[Unit]
Description=rinetd is an TCP/UDP redirection server 
After=network.target

[Service]
Type=simple
Restart=always
ExecStart=${var_cur}/rinetd
WorkingDirectory=${var_cur}
PrivateTmp=true

[Install]
WantedBy=multi-user.target
Alias=rinetd.service

EOF

# uninstall old service
sudo -u root systemctl disable ${var_service_filename} || true
sudo -u root rm -f /usr/lib/systemd/system/${var_service_filename} || true
sudo -u root rm -f /etc/systemd/system/${var_service_filename} || true

sudo -u root mv ${var_dst_path} /usr/lib/systemd/system/${var_service_filename}
sudo -u root systemctl daemon-reload
sudo -u root systemctl enable ${var_service_filename}
sudo -u root systemctl start ${var_service_filename}

echo "success install rinetd service"
