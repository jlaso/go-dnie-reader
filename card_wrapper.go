package main

import (
	"fmt"
	"github.com/ebfe/scard"
)

type CardWrapper struct {
	Card     *scard.Card
	SW1      byte
	SW2      byte
	Response []byte
}

var context *scard.Context
var err error

func (cw *CardWrapper) Connect() error {
	// Establish a PC/SC context
	context, err = scard.EstablishContext()
	if err != nil {
		fmt.Println("Error EstablishContext:", err)
		return err
	}

	// List available readers
	readers, err := context.ListReaders()
	if err != nil {
		fmt.Println("Error ListReaders:", err)
		return err
	}

	// Use the first reader
	reader := readers[0]
	fmt.Println("Using reader:", reader)

	// Connect to the card
	card, err := context.Connect(reader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		fmt.Println("Error Connect:", err)
		return err
	}
	cw.Card = card

	return nil
}

func (cw *CardWrapper) Disconnect() {
	// Release the PC/SC context (when needed)
	defer context.Release()
	// Disconnect (when needed)
	defer cw.Card.Disconnect(scard.LeaveCard)
}

/*
	since %X doesn't fill in leading left zeroes...
 */
func hex(b byte) string {
	lo := byte(b % 16)
	hi := byte(b / 16)
	return fmt.Sprintf("%X%X", hi, lo)
}

func (cw *CardWrapper) ValidSW(sw1valids ...byte) error {
	for _, n := range sw1valids {
		if n == cw.SW1 {
			return nil
		}
	}
	found := uint16(cw.SW2) + uint16(cw.SW1)*256
	if found == 0x9000 {
		// Ejecución correcta.
		return nil
	}
	if found == 0x6283 {
		return fmt.Errorf("%s", "El EF o DF seleccionados están invalidados")
	}
	if found == 0x6581 {
		return fmt.Errorf("%s", "Error de memoria")
	}
	if found == 0x6700 {
		return fmt.Errorf("%s", "Longitud incorrecta (el campo Lc no es correcto)")
	}
	if found == 0x6982 {
		return fmt.Errorf("%s", "Condiciones de seguridad no satisfechas")
	}
	if found == 0x6985 {
		return fmt.Errorf("%s", "Condiciones de uso no satisfechas")
	}
	if found == 0x6986 {
		return fmt.Errorf("%s", "Comando no permitido (no existe un EF seleccionado)")
	}
	if found == 0x6A82 {
		return fmt.Errorf("%s", "Fichero no encontrado")
	}
	if found == 0x6A86 {
		return fmt.Errorf("%s", "Parámetros P1-P2 incorrectos")
	}
	if found == 0x6B00 {
		return fmt.Errorf("%s", "Parámetros incorrectos (el offset está fuera del EF)")
	}
	if cw.SW1 == 0x6C {
		return fmt.Errorf("longitud incorrecta (%d es el valor correcto)", cw.SW2)
	}
	return fmt.Errorf("unexpected %X SW found", found)
}

func (cw *CardWrapper) sendCommand(cmd []byte, validSW1 ...byte) error {
	resp, err := cw.Card.Transmit(cmd)
	if err != nil {
		return err
	}
	if len(resp) >= 2 {
		cw.SW1 = resp[len(resp)-2]
		cw.SW2 = resp[len(resp)-1]
	}
	if err = cw.ValidSW(validSW1...); err != nil {
		return err
	}
	cw.Response = resp[:len(resp)-2]
	return nil
}

func (cw *CardWrapper) CardAccess() (string, []byte, error) {
	if err := cw.SelectEF(0x1c01); err != nil {
		return "", nil, err
	}
	if err = cw.GetResponse(); err != nil {
		return "", nil, err
	}
	var ci string
	ci = ""
	for _, i := range cw.Response {
		ci += hex(i)
	}
	var data []byte
	_len := uint16(cw.Response[7])*256 + uint16(cw.Response[8])
	if data, err = cw.ReadBinary(0x0000, _len); err != nil {
		return "", nil, err
	}
	return ci, data, nil
}

func (cw *CardWrapper) IDEsp() (string, error) {
	if err := cw.SelectEF(0x0600); err != nil {
		return "", err
	}
	if err = cw.GetResponse(); err != nil {
		return "", err
	}
	// fci := cw.Response
	var data []byte
	_len := uint16(cw.Response[7])*256 + uint16(cw.Response[8])
	if data, err = cw.ReadBinary(0x0000, _len); err != nil {
		return "", err
	}
	return string(data), nil
}

func (cw *CardWrapper) GetChipInfo() (string, error) {
	cmd := []byte{0x90, 0xb8, 0x00, 0x00, 0x07}
	if err := cw.sendCommand(cmd); err != nil {
		return "", err
	}
	var ci string
	ci = ""
	for _, i := range cw.Response {
		ci += hex(i)
	}
	return ci, nil
}

func (cw *CardWrapper) SelectEF(ef uint16) error {
	cmd := []byte{0x00, 0xa4, 0x00, 0x00, 0x02, 0x00, 0x00}
	cmd[5] = byte(ef % 256)
	cmd[6] = byte(ef / 256)
	return cw.sendCommand(cmd, 0x61)
}

func (cw *CardWrapper) GetResponse() error {
	cmd := []byte{0x00, 0xc0, 0x00, 0x00, 0x00}
	cmd[4] = cw.SW2
	return cw.sendCommand(cmd)
}

func (cw *CardWrapper) ReadBinary(offset uint16, _len uint16) ([]byte, error) {
	var err error
	var off uint16
	var this_len byte
	result := make([]byte, _len)
	off = 0
	for _len > 0 {
		if _len > 255 {
			this_len = 255
		} else {
			this_len = byte(_len)
		}
		cmd := []byte{0x00, 0xb0, 0x00, 0x00, 0x00}
		cmd[2] = byte(offset % 256)
		cmd[3] = byte(offset / 256)
		cmd[4] = this_len
		if err = cw.sendCommand(cmd); err != nil {
			return nil, err
		}
		for i, n := range cw.Response {
			result[off+uint16(i)] = n
		}
		off += uint16(this_len)
		offset += uint16(this_len)
		_len -= uint16(this_len)
	}
	return result, nil
}

func (cw *CardWrapper) PrettyPrint(area string) {
	fmt.Printf("resp %s: \n", area)
	fmt.Print("[")
	for i := 0; i < len(cw.Response); i++ {
		fmt.Print(hex(cw.Response[i]))
	}
	fmt.Print(" ]  ")
	for i := 0; i < len(cw.Response); i++ {
		fmt.Printf("%c", cw.Response[i])
	}
	fmt.Printf(" |SW1:0x%X|SW2:0x%X|\n", cw.SW1, cw.SW2)
}

func to_char(b byte) string {
	if b >=32 && b < 128 {
		return fmt.Sprintf("%c", b)
	}
	return " "
}

func PrettyPrint(data []byte) {
	s := ""
	var i int
	for i = 0; i < len(data); i++ {
		if i%16 == 0 {
			fmt.Println("    ", s)
			s = ""
		}
		fmt.Print(hex(data[i]))
		s += to_char(data[i])
	}
	for j:= i; j< 16; j++{
		fmt.Print(" ")
	}
	fmt.Println("    ", s, "\n")
}
