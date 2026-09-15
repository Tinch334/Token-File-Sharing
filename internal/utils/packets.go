package utils


import (
    "fmt"
    "slices"
    "errors"
    "net"
    "io"
    "encoding/binary"

    "github.com/Tinch334/Token-File-Sharing/internal/constants"
)


// ReadData reads all of the packet's data blocks.
func ReadData(conn net.Conn) ([][]byte, error) {
    // Read data blocks until CONTINUATION_BYTE_NO.
    var data [][]byte
    for {
        // Read packet data size.
        packetSizeBytes := make([]byte, 4)
        if _, err := io.ReadFull(conn, packetSizeBytes); err != nil {
            if err == io.ErrUnexpectedEOF {
                SendError("Truncated block size", conn)
            }
            return nil, fmt.Errorf("Error reading block size: %w", err)
        }

        // Read packet data.
        blockSize := binary.BigEndian.Uint32(packetSizeBytes)
        block := make([]byte, blockSize)
        if _, err := io.ReadFull(conn, block); err != nil {
            if err == io.ErrUnexpectedEOF {
                SendError("Truncated block data", conn)
            }
            return nil, fmt.Errorf("Error reading block data: %w", err)
        }

        // Store read packet data.
        data = append(data, block)

        // Read continuation indicator.
        cIndicator := make([]byte, 1)
        if _, err := io.ReadFull(conn, cIndicator); err != nil {
            if err == io.ErrUnexpectedEOF {
                SendError("Truncated continuation data", conn)
            }
            return nil, fmt.Errorf("Error continuation bytes: %w", err)
        }

        fmt.Printf("data:%X  blockSize:%d  block:%X  cont:%X\n",
            data, blockSize, block, cIndicator)

        // Check if there's more data to be read.
        if cIndicator[0] == constants.CONTINUATION_BYTE_NO {
            break
        }
    }

    return data, nil
}

// MakePacket returns a properly formatted packet with the given header and data.
func MakePacket(header byte, dataList [][]byte) ([]byte, error) {
    var res []byte
    // Add header.
    res = []byte{header}

    for i, block := range dataList {
        // Get data length and check it's validity.
        if len(block) > constants.MAX_DATA_SIZE {
            return nil, errors.New("Data exceeded maximum data size")
        }

        lenBuf := make([]byte, 4)
        binary.BigEndian.PutUint32(lenBuf, uint32(len(block)))

        // Grow slice to fit block size, block and continuation indicator, append data.
        res = slices.Grow(res, constants.DATA_BLOCK_SIZE + len(block) + 1)
        res = append(res, lenBuf...)
        res = append(res, block...)
    
        // Set continuation byte.
        if i < len(dataList) - 1 {
            res = append(res, constants.CONTINUATION_BYTE_YES)
        } else {
            res = append(res, constants.CONTINUATION_BYTE_NO)
        }
    }

    return res, nil
}


// MakeTokenPacket is a wrapper around "MakePacket" that makes a packet with the specified token, if the given token is an empty,
// string then auth token length will be set to zero to indicate no token is present.
func MakeTokenPacket(token string, header byte, dataList [][]byte) ([]byte, error) {
    lenBuf := make([]byte, 2)
    binary.BigEndian.PutUint16(lenBuf, uint16(len(token)))
    
    args := [][]byte{lenBuf}

    if (len(token) > 0) {
        args = append(args, []byte(token))    
    }

    args = append(args, dataList...)

    return MakePacket(header, args)
}


// MakeErrorPacket is a wrapper around "MakePacket" that makes an error packet.
func MakeErrorPacket (err string) ([]byte, error) {
    return MakePacket(constants.ERR, [][]byte{[]byte(err)})
}


// SendPacket sends the given data to the connection.
func SendPacket(data []byte, conn net.Conn) error {
    _, err := conn.Write(data)
    return err
}


// SendError sends an ERR packet to conn with the given error string, the error will not be returned.
// An error is only returned in case of a send error.
func SendError(errorMsg string, conn net.Conn) error {
    // Create a slice of slices with the error message.
    packet, err := MakePacket(constants.ERR, [][]byte{[]byte(errorMsg)})
    if err != nil {
        return fmt.Errorf("Error building error packet: %w\n", err)
    }

    if  _, err = conn.Write(packet); err != nil {
        return fmt.Errorf("Error building error packet: %w\n", err)
    }

    return err
}