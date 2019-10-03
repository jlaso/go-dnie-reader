# Super Smart Card(sscard)

Super Smart Card API on top of scard(pcsc handler) with apdu commands.

## Builtin APDU 

- DNIe card (public data)

## TODO

``` bash
# Linux: install pcsc library
sudo apt-get install pcscd

# goget dependencies
go get github.com/gogetth/sscard

# build example
go build -o read_dnie *.go

./read_dnie

# run without building
go run *.go
```

## References

- [Thank you for original sscard work from Napat](https://github.com/Napat/sscard)
- [Thank you for sscard wrapper from gogetth](https://github.com/gogetth/sscard)
- [PCSC in golang](https://ludovicrousseau.blogspot.fr/2016/09/pcsc-sample-in-go.html)
- [APDU command for DNIe](https://www.dnielectronico.es/PDFs/Manual_de_Comandos_para_Desarrolladores_102.pdf)
