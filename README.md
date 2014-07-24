# UDP Transport

To test with a local loop :

    sudo modprobe snd-aloop
    mplayer -loop 0 -ao alsa:device=hw=1.1.1 -srate 48000 input.wav
    go-broadcast udpclient --alsa-device=hw:1,0,1 --udp-target=localhost:9090 -log-debug -http-bind=:9001
    go-broadcast udpserver -log-debug -http-bind=:9000

# Backup

To test with smaller files :

    go-broadcast backup --file-duration=1m /tmp/records

# Loopback

    sudo modprobe snd-aloop
    mplayer -loop 0 -ao alsa:device=hw=1.1.1 -srate 44110 input.wav
    go-broadcast loopback --input-device=hw:1,0,1 --output-device=default

# Requirements to build

    sudo apt-get install libvorbis-dev libasound2-dev libopus-dev
