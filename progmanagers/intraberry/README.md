# Intraberry

This program manager allows to run only one binary called program from sdcard (first partition) on raspberry.
Program manager interface is http on port 4242. This is totally unsafe, ok for developing on safe intranet environment

This allows fast iteration on development

## Build

## Install
Copy initramfs to fat partition

## Configuration

All configurations are on first sdcard partition mounted under dir **/intraberry**

| File name | Description | Default |
|-----------|--------------------|---------|
| tz.txt | timezon string on file| Europe/Helsinki|
| host.txt | hostname | intraberry |
| eth0.ip | ethernet IP number with mask | use DHCP |
| eth0.gw | ethernet gateway | use DHCP |
| eth0.ns | nameservers (DNS) on IP per line | use DHCP |
| ntp.txt | ntp server IPs one IP per line | use finnish NTP servers |

## API

| Endpoint    | method   | description |
|-------------|----------|-------------|
| **/stdout** | HTTP GET | what program prints out |
| **/stderr** | HTTP GET |
| **/restart** | HTTP GET | restarts software to persisted |
| **/reboot**  | HTTP GET |reboots
| **/tmpprog** | HTTP POST | upload program to /tmp/program and runs it without access to sdcard|
| **/prog**    | HTTP POST | uploads and replaces program |

TODO thinking
- just cancel context that re-starts
- changeable program name
- checking that elf is arm64
