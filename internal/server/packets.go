package server


import (
    "fmt"
    "slices"
    "errors"
    "net"
    "io"
    "encoding/binary"

    "github.com/Tinch334/Token-File-Sharing/internal/constants"
)


// readData reads all of the packet's data blocks.
func readData(conn net.Conn) ([][]byte, error) {
    // Read data blocks until ADDITIONAL_BLOCK_NO.
    var data [][]byte
    for {
        // Read packet data size.
        packetSizeBytes := make([]byte, 4)
        if _, err := io.ReadFull(conn, packetSizeBytes); err != nil {
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated block size", conn)
            }
            return nil, fmt.Errorf("Error reading block size: %w", err)
        }

        // Read packet data.
        blockSize := binary.BigEndian.Uint32(packetSizeBytes)
        block := make([]byte, blockSize)
        if _, err := io.ReadFull(conn, block); err != nil {
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated block data", conn)
            }
            return nil, fmt.Errorf("Error reading block data: %w", err)
        }

        // Store read packet data.
        data = append(data, block)

        // Read continuation indicator.
        cIndicator := make([]byte, 1)
        if _, err := io.ReadFull(conn, cIndicator); err != nil {
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated continuation data", conn)
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

// makePacket returns a properly formatted packet with the given header and data.
func makePacket(header byte, dataList [][]byte) ([]byte, error) {
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


// makeErrorPacket is a wrapper around "makePacket" that makes an error packet.
func makeErrorPacket (err string) ([]byte, error) {
    return makePacket(constants.ERR, [][]byte{[]byte(err)})
}


// sendPacket sends the given data to the connection.
func sendPacket(data []byte, conn net.Conn) error {
    _, err := conn.Write(data)
    return err
}


// sendError sends an ERR packet to conn with the given error string, the error will not be returned.
// An error is only returned in case of a send error.
func sendError(errorMsg string, conn net.Conn) error {
    // Create a slice of slices with the error message.
    packet, err := makePacket(constants.ERR, [][]byte{[]byte(errorMsg)})
    if err != nil {
        return fmt.Errorf("Error building error packet: %w\n", err)
    }

    if  _, err = conn.Write(packet); err != nil {
        return fmt.Errorf("Error building error packet: %w\n", err)
    }

    return err
}