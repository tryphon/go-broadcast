# UDP Transport

To test with a local loop :

    sudo modprobe snd-aloop
    mplayer -loop 0 -ao alsa:device=hw=1.1.1 -srate 48000 input.wav
    go-broadcast udpclient --alsa-device=hw:1,0,1 localhost:7890
    go-broadcast udpserver :7890
