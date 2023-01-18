# LibMITM
Dead simple gVisor netstack powered packet processing library on Android

## Build requirements
* JDK
* Android SDK
* Go
* gomobile

## Build instructions
1. `git clone [repo] && cd [repo]`
2. `gomobile init`
3. `gomobile bind -v -androidapi 28 -ldflags='-s -w' ./`

## Credit
[tun2socks](https://github.com/xjasonlyu/tun2socks),
[Matsuri](https://github.com/MatsuriDayo/Matsuri) for helps me understanding gVisor netstack