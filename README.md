# Chip 8

![Screenshot of the emulator displaying the IBM logo.](ibmlogo.png)

## Current State

For now, original & schip-1.1 (partially for now) are supported. Support for Xo chip is planned. The emulator passes all the tests (other than the display wait quirk for the original spec) from [Timendus's suite](https://github.com/Timendus/chip8-test-suite).

## Usage

Run with the command `./GoCh8 --rom=path/to/rom`. Upon starting, the keyboard mapping will be printed to the console.

### Optional CLI arguments

1. --color: Specifies the color scheme. Black, yellow and green color schemes are available. Default is Green.
2. --spec: Specifies the specification of Chip 8 to emulate. Original, Super and Xo are available. Default is Original.
3. --speed: Specifies an integer speed multiplier for the emulation. Original is 1.

## Thanks to

- [Timendus's Chip 8 test suite](https://github.com/Timendus/chip8-test-suite)
- [Guide to making a CHIP-8 emulator](https://tobiasvl.github.io/blog/write-a-chip-8-emulator/)
- [CHIP-8 Variant Opcode Table](https://chip8.gulrak.net)
- The amazing people of EmuDev Discord.
