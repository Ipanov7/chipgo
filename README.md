Chipgo is a Golang implementation of the CHIP-8 processor. Technical specs were taken from Cowgod's article [here](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#00EE)

# TODO
- [ ] Input
- [ ] Audio
- [ ] Replace virtual stack with in-memory implementation
- [ ] Use in-memory buffer to store sprites info before rendering

# Usage
There are 2 programs in this project:
- disassembler (to debug binary ROMs)
- chipgo

Just compile the `chipgo` program and then run:

```sh
$ ./chipgo /path/to/rom [--debug]
```

The `debug` flag allow to run a step-by-step execution of the ROM. Instructions can be cycled by pressing `ENTER`.
