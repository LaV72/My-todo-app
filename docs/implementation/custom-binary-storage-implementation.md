# Custom Binary Storage Implementation

## Overview

This document details the design and implementation of a **custom binary file format** for storing Quest Todo data with built-in indexes for fast queries.

**Approach**: Multi-section binary file format with header, index section, and data section.

**Philosophy**: Build a mini-database format from scratch to understand how databases work internally.

**Trade-offs**:
- ✅ Educational (learn file formats, binary encoding, index structures)
- ✅ Fast queries (indexes stored with data)
- ✅ Flexible loading (can load just indexes, or all data)
- ✅ Extensible (easy to add new sections)
- ✅ Production-ready potential
- ❌ Complex implementation (~800 lines)
- ❌ Custom format (need good documentation)
- ❌ Binary (not human-readable)

**Best For**: Learning database internals, production applications, large datasets (10k-1M records)

## File Format Overview

### Binary File Structure

```
┌─────────────────────────────────────────────────────────────┐
│ FILE HEADER (256 bytes fixed)                               │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Magic Number    [4 bytes]   "QTDO"                      │ │
│ │ Format Version  [4 bytes]   1, 2, 3, ...                │ │
│ │ Flags           [4 bytes]   Compression, encryption, etc │ │
│ │ Created At      [8 bytes]   Unix timestamp              │ │
│ │ Modified At     [8 bytes]   Unix timestamp              │ │
│ │ NumTasks        [4 bytes]   Count of tasks              │ │
│ │ NumObjectives   [4 bytes]   Count of objectives         │ │
│ │ NumCategories   [4 bytes]   Count of categories         │ │
│ │ IndexOffset     [8 bytes]   Byte offset to index section│ │
│ │ DataOffset      [8 bytes]   Byte offset to data section │ │
│ │ WALOffset       [8 bytes]   Byte offset to WAL section  │ │
│ │ Checksum        [4 bytes]   CRC32 of entire file        │ │
│ │ Reserved        [196 bytes] For future use              │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ INDEX SECTION (variable length)                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Index Count     [4 bytes]   Number of indexes           │ │
│ │                                                         │ │
│ │ [Index 1: Status Index]                                │ │
│ │   Type          [1 byte]    1 = Status                 │ │
│ │   Size          [4 bytes]   Size in bytes              │ │
│ │   Data          [N bytes]   Serialized index           │ │
│ │                                                         │ │
│ │ [Index 2: Priority Index]                              │ │
│ │   Type          [1 byte]    2 = Priority               │ │
│ │   Size          [4 bytes]   Size in bytes              │ │
│ │   Data          [N bytes]   Serialized index           │ │
│ │                                                         │ │
│ │ [Index 3: Category Index]                              │ │
│ │ [Index 4: Tag Index]                                   │ │
│ │ [Index 5: Word Index]                                  │ │
│ │ [Index 6: CreatedAt Sorted]                            │ │
│ │ ...                                                    │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ DATA SECTION (variable length)                              │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ [Task Block 1]                                          │ │
│ │   Length        [4 bytes]   Block size                 │ │
│ │   ID            [36 bytes]  UUID string                │ │
│ │   Flags         [1 byte]    Deleted, etc.              │ │
│ │   Data          [N bytes]   Serialized task            │ │
│ │   Checksum      [4 bytes]   CRC32 of block             │ │
│ │                                                         │ │
│ │ [Task Block 2]                                          │ │
│ │   ...                                                   │ │
│ │                                                         │ │
│ │ [Objective Blocks]                                      │ │
│ │   ...                                                   │ │
│ │                                                         │ │
│ │ [Category Blocks]                                       │ │
│ │   ...                                                   │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ WAL SECTION (optional, variable length)                     │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ [Operation 1]                                           │ │
│ │   Timestamp     [8 bytes]   When operation occurred    │ │
│ │   OpType        [1 byte]    CREATE, UPDATE, DELETE     │ │
│ │   DataLength    [4 bytes]   Length of operation data   │ │
│ │   Data          [N bytes]   Serialized operation       │ │
│ │                                                         │ │
│ │ [Operation 2]                                           │ │
│ │   ...                                                   │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

#### 1. Fixed-Size Header (256 bytes)
- Easy to seek and read
- Version field for format evolution
- Offsets allow sections to be reordered
- Reserved space for future fields

#### 2. Self-Describing Indexes
- Each index has type ID
- Can skip unknown index types (forward compatibility)
- Indexes stored together for locality

#### 3. Length-Prefixed Blocks
- Each data block has its length
- Enables sequential scanning
- Supports variable-size records

#### 4. Checksums
- File-level checksum (header)
- Block-level checksums (data)
- Detect corruption early

#### 5. Optional WAL Section
- Append-only log for durability
- Can replay after crash
- Periodic compaction into main sections

## Binary Encoding Details

### Data Type Encoding

We'll use **little-endian** encoding (Intel/AMD standard):

```go
// Encoding primitives
func writeUint32(w io.Writer, v uint32) error {
    buf := make([]byte, 4)
    binary.LittleEndian.PutUint32(buf, v)
    _, err := w.Write(buf)
    return err
}

func readUint32(r io.Reader) (uint32, error) {
    buf := make([]byte, 4)
    if _, err := io.ReadFull(r, buf); err != nil {
        return 0, err
    }
    return binary.LittleEndian.Uint32(buf), nil
}

func writeString(w io.Writer, s string) error {
    // Write length prefix (4 bytes)
    if err := writeUint32(w, uint32(len(s))); err != nil {
        return err
    }
    // Write string bytes
    _, err := w.Write([]byte(s))
    return err
}

func readString(r io.Reader) (string, error) {
    // Read length
    length, err := readUint32(r)
    if err != nil {
        return "", err
    }
    // Read string bytes
    buf := make([]byte, length)
    if _, err := io.ReadFull(r, buf); err != nil {
        return "", err
    }
    return string(buf), nil
}
```

### Type Sizes

| Type | Size | Encoding |
|------|------|----------|
| uint8 | 1 byte | Direct |
| uint32 | 4 bytes | Little-endian |
| uint64 | 8 bytes | Little-endian |
| int32 | 4 bytes | Little-endian |
| int64 | 8 bytes | Little-endian |
| bool | 1 byte | 0 or 1 |
| string | 4 + N bytes | Length prefix + UTF-8 |
| time.Time | 8 bytes | Unix timestamp (int64) |
| []string | 4 + (4+N)*M bytes | Count + length-prefixed strings |

## Header Section

### Header Structure

```go
const (
    MagicNumber = 0x4F445451  // "QTDO" in hex
    CurrentVersion = 1
    HeaderSize = 256
)

type FileHeader struct {
    Magic         uint32  // Must be MagicNumber
    Version       uint32  // Format version
    Flags         uint32  // Bit flags
    CreatedAt     int64   // Unix timestamp
    ModifiedAt    int64   // Unix timestamp
    NumTasks      uint32  // Count of tasks
    NumObjectives uint32  // Count of objectives
    NumCategories uint32  // Count of categories
    IndexOffset   uint64  // Byte offset to index section
    DataOffset    uint64  // Byte offset to data section
    WALOffset     uint64  // Byte offset to WAL section (0 if none)
    Checksum      uint32  // CRC32 of file (excluding this field)
    Reserved      [196]byte // For future use
}
```

### Header Flags

```go
const (
    FlagCompressed   = 1 << 0  // Data is compressed
    FlagEncrypted    = 1 << 1  // Data is encrypted
    FlagHasWAL       = 1 << 2  // WAL section present
    FlagIndexesOnly  = 1 << 3  // Only indexes, no data (for metadata)
)
```

### Writing Header

```go
func (h *FileHeader) Write(w io.Writer) error {
    // Create buffer for header
    buf := new(bytes.Buffer)

    // Write all fields
    binary.Write(buf, binary.LittleEndian, h.Magic)
    binary.Write(buf, binary.LittleEndian, h.Version)
    binary.Write(buf, binary.LittleEndian, h.Flags)
    binary.Write(buf, binary.LittleEndian, h.CreatedAt)
    binary.Write(buf, binary.LittleEndian, h.ModifiedAt)
    binary.Write(buf, binary.LittleEndian, h.NumTasks)
    binary.Write(buf, binary.LittleEndian, h.NumObjectives)
    binary.Write(buf, binary.LittleEndian, h.NumCategories)
    binary.Write(buf, binary.LittleEndian, h.IndexOffset)
    binary.Write(buf, binary.LittleEndian, h.DataOffset)
    binary.Write(buf, binary.LittleEndian, h.WALOffset)
    binary.Write(buf, binary.LittleEndian, h.Checksum)
    buf.Write(h.Reserved[:])

    // Ensure exactly 256 bytes
    if buf.Len() != HeaderSize {
        return fmt.Errorf("header size mismatch: got %d, want %d", buf.Len(), HeaderSize)
    }

    _, err := w.Write(buf.Bytes())
    return err
}
```

### Reading Header

```go
func ReadHeader(r io.Reader) (*FileHeader, error) {
    h := &FileHeader{}

    // Read all fields
    if err := binary.Read(r, binary.LittleEndian, &h.Magic); err != nil {
        return nil, err
    }

    // Verify magic number
    if h.Magic != MagicNumber {
        return nil, fmt.Errorf("invalid magic number: got 0x%X, want 0x%X", h.Magic, MagicNumber)
    }

    // Read remaining fields
    binary.Read(r, binary.LittleEndian, &h.Version)
    binary.Read(r, binary.LittleEndian, &h.Flags)
    binary.Read(r, binary.LittleEndian, &h.CreatedAt)
    binary.Read(r, binary.LittleEndian, &h.ModifiedAt)
    binary.Read(r, binary.LittleEndian, &h.NumTasks)
    binary.Read(r, binary.LittleEndian, &h.NumObjectives)
    binary.Read(r, binary.LittleEndian, &h.NumCategories)
    binary.Read(r, binary.LittleEndian, &h.IndexOffset)
    binary.Read(r, binary.LittleEndian, &h.DataOffset)
    binary.Read(r, binary.LittleEndian, &h.WALOffset)
    binary.Read(r, binary.LittleEndian, &h.Checksum)
    io.ReadFull(r, h.Reserved[:])

    return h, nil
}
```

## Index Section

### Index Types

```go
const (
    IndexTypeStatus       = 1
    IndexTypePriority     = 2
    IndexTypeCategory     = 3
    IndexTypeTag          = 4
    IndexTypeDeadlineType = 5
    IndexTypeWord         = 6
    IndexTypePrioritySorted = 7
    IndexTypeCreatedAtSorted = 8
)
```

### Index Entry Format

Each index entry:
```
┌──────────────────────────────┐
│ Type     [1 byte]            │
│ Size     [4 bytes]           │
│ Data     [N bytes]           │
└──────────────────────────────┘
```

### Status Index Encoding

Format: `map[TaskStatus][]string` → list of (status, task IDs)

```
┌──────────────────────────────────────────┐
│ NumStatuses    [4 bytes]                 │
│                                          │
│ [Entry 1]                                │
│   Status       [1 byte]   (0=active)     │
│   NumTaskIDs   [4 bytes]                 │
│   TaskID1      [36 bytes] UUID string    │
│   TaskID2      [36 bytes] UUID string    │
│   ...                                    │
│                                          │
│ [Entry 2]                                │
│   Status       [1 byte]   (1=completed)  │
│   NumTaskIDs   [4 bytes]                 │
│   TaskID1      [36 bytes]                │
│   ...                                    │
└──────────────────────────────────────────┘
```

**Encoding**:
```go
func encodeStatusIndex(w io.Writer, index map[models.TaskStatus]map[string]bool) error {
    // Write number of statuses
    writeUint32(w, uint32(len(index)))

    for status, taskIDs := range index {
        // Write status (as uint8)
        w.Write([]byte{uint8(status)})

        // Write task ID count
        writeUint32(w, uint32(len(taskIDs)))

        // Write each task ID (fixed 36 bytes for UUID)
        for taskID := range taskIDs {
            w.Write([]byte(fmt.Sprintf("%-36s", taskID)))  // Pad to 36 bytes
        }
    }
    return nil
}
```

**Decoding**:
```go
func decodeStatusIndex(r io.Reader) (map[models.TaskStatus]map[string]bool, error) {
    index := make(map[models.TaskStatus]map[string]bool)

    // Read number of statuses
    numStatuses, _ := readUint32(r)

    for i := 0; i < int(numStatuses); i++ {
        // Read status
        statusByte := make([]byte, 1)
        r.Read(statusByte)
        status := models.TaskStatus(statusByte[0])

        // Read task ID count
        numTaskIDs, _ := readUint32(r)

        // Read task IDs
        taskIDs := make(map[string]bool)
        for j := 0; j < int(numTaskIDs); j++ {
            idBuf := make([]byte, 36)
            io.ReadFull(r, idBuf)
            taskID := strings.TrimSpace(string(idBuf))
            taskIDs[taskID] = true
        }

        index[status] = taskIDs
    }

    return index, nil
}
```

### Tag Index Encoding (Inverted Index)

Format: `map[string][]string` → tag → task IDs

```
┌──────────────────────────────────────────┐
│ NumTags        [4 bytes]                 │
│                                          │
│ [Entry 1]                                │
│   TagLength    [4 bytes]                 │
│   Tag          [N bytes]  "urgent"       │
│   NumTaskIDs   [4 bytes]                 │
│   TaskID1      [36 bytes]                │
│   TaskID2      [36 bytes]                │
│   ...                                    │
│                                          │
│ [Entry 2]                                │
│   TagLength    [4 bytes]                 │
│   Tag          [N bytes]  "bug"          │
│   NumTaskIDs   [4 bytes]                 │
│   TaskID1      [36 bytes]                │
│   ...                                    │
└──────────────────────────────────────────┘
```

### Word Index Encoding

Same format as Tag Index (both are inverted indexes).

### Sorted Index Encoding

Format: Sorted list of task IDs

```
┌──────────────────────────────────────────┐
│ NumTaskIDs     [4 bytes]                 │
│ TaskID1        [36 bytes]                │
│ TaskID2        [36 bytes]                │
│ TaskID3        [36 bytes]                │
│ ...                                      │
└──────────────────────────────────────────┘
```

## Data Section

### Data Block Format

Each record (task, objective, category) is a **data block**:

```
┌──────────────────────────────────────────┐
│ BlockType      [1 byte]   1=Task         │
│ Length         [4 bytes]  Total size     │
│ ID             [36 bytes] UUID           │
│ Flags          [1 byte]   Deleted, etc.  │
│ DataLength     [4 bytes]  Payload size   │
│ Data           [N bytes]  Serialized     │
│ Checksum       [4 bytes]  CRC32          │
└──────────────────────────────────────────┘
```

**Block Types**:
```go
const (
    BlockTypeTask      = 1
    BlockTypeObjective = 2
    BlockTypeCategory  = 3
)
```

**Block Flags**:
```go
const (
    BlockFlagDeleted   = 1 << 0  // Soft delete
    BlockFlagCompressed = 1 << 1  // Data is compressed
)
```

### Task Serialization

For the data payload, we can use:

**Option A: JSON** (simple, debuggable)
```go
func encodeTask(task *models.Task) ([]byte, error) {
    return json.Marshal(task)
}
```

**Option B: MessagePack** (compact, fast)
```go
func encodeTask(task *models.Task) ([]byte, error) {
    return msgpack.Marshal(task)
}
```

**Option C: Custom Binary** (maximum control)
```go
func encodeTask(w io.Writer, task *models.Task) error {
    writeString(w, task.ID)
    writeString(w, task.Title)
    writeString(w, task.Description)
    writeUint32(w, uint32(task.Priority))
    // ... all fields
}
```

**Recommendation**: Start with **MessagePack** (good balance of simplicity and efficiency).

### Writing Data Block

```go
func writeTaskBlock(w io.Writer, task *models.Task) error {
    // Encode task data
    data, err := msgpack.Marshal(task)
    if err != nil {
        return err
    }

    // Calculate checksum
    checksum := crc32.ChecksumIEEE(data)

    // Write block
    w.Write([]byte{BlockTypeTask})
    writeUint32(w, uint32(1 + 4 + 36 + 1 + 4 + len(data) + 4))  // Total length
    w.Write([]byte(fmt.Sprintf("%-36s", task.ID)))  // Fixed-size ID
    w.Write([]byte{0})  // Flags
    writeUint32(w, uint32(len(data)))
    w.Write(data)
    writeUint32(w, checksum)

    return nil
}
```

### Reading Data Block

```go
func readDataBlock(r io.Reader) (*DataBlock, error) {
    block := &DataBlock{}

    // Read block type
    typeBuf := make([]byte, 1)
    r.Read(typeBuf)
    block.Type = typeBuf[0]

    // Read length
    block.Length, _ = readUint32(r)

    // Read ID
    idBuf := make([]byte, 36)
    io.ReadFull(r, idBuf)
    block.ID = strings.TrimSpace(string(idBuf))

    // Read flags
    flagsBuf := make([]byte, 1)
    r.Read(flagsBuf)
    block.Flags = flagsBuf[0]

    // Read data
    dataLength, _ := readUint32(r)
    block.Data = make([]byte, dataLength)
    io.ReadFull(r, block.Data)

    // Read checksum
    block.Checksum, _ = readUint32(r)

    // Verify checksum
    actualChecksum := crc32.ChecksumIEEE(block.Data)
    if actualChecksum != block.Checksum {
        return nil, fmt.Errorf("checksum mismatch for block %s", block.ID)
    }

    return block, nil
}
```

## WAL Section (Optional)

### WAL Entry Format

```
┌──────────────────────────────────────────┐
│ Timestamp      [8 bytes]  Unix nano      │
│ OpType         [1 byte]   1=CREATE, etc. │
│ EntityType     [1 byte]   1=Task, etc.   │
│ DataLength     [4 bytes]                 │
│ Data           [N bytes]  Serialized     │
│ Checksum       [4 bytes]  CRC32          │
└──────────────────────────────────────────┘
```

**Operation Types**:
```go
const (
    OpTypeCreate = 1
    OpTypeUpdate = 2
    OpTypeDelete = 3
)
```

### Append to WAL

```go
func (s *CustomStorage) appendToWAL(opType uint8, entityType uint8, data []byte) error {
    // Open WAL file in append mode
    f, err := os.OpenFile(s.walPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    // Write WAL entry
    binary.Write(f, binary.LittleEndian, time.Now().UnixNano())
    f.Write([]byte{opType, entityType})
    writeUint32(f, uint32(len(data)))
    f.Write(data)
    checksum := crc32.ChecksumIEEE(data)
    writeUint32(f, checksum)

    // Sync to disk (durability)
    return f.Sync()
}
```

### Replay WAL

```go
func (s *CustomStorage) replayWAL() error {
    f, err := os.Open(s.walPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil  // No WAL
        }
        return err
    }
    defer f.Close()

    for {
        // Read WAL entry
        var timestamp int64
        if err := binary.Read(f, binary.LittleEndian, &timestamp); err != nil {
            if err == io.EOF {
                break  // End of WAL
            }
            return err
        }

        opType, _ := readUint8(f)
        entityType, _ := readUint8(f)
        dataLength, _ := readUint32(f)
        data := make([]byte, dataLength)
        io.ReadFull(f, data)
        checksum, _ := readUint32(f)

        // Verify checksum
        if crc32.ChecksumIEEE(data) != checksum {
            return fmt.Errorf("WAL entry corrupt at offset %d", offset)
        }

        // Apply operation
        switch opType {
        case OpTypeCreate:
            s.applyCreate(entityType, data)
        case OpTypeUpdate:
            s.applyUpdate(entityType, data)
        case OpTypeDelete:
            s.applyDelete(entityType, data)
        }
    }

    return nil
}
```

## Loading Strategies

### Strategy 1: Full Load (Simple)

Load everything into memory at startup:

```go
func (s *CustomStorage) Load(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()

    // 1. Read header
    header, err := ReadHeader(f)
    if err != nil {
        return err
    }

    // 2. Seek to index section
    f.Seek(int64(header.IndexOffset), io.SeekStart)

    // 3. Load all indexes
    if err := s.loadIndexes(f); err != nil {
        return err
    }

    // 4. Seek to data section
    f.Seek(int64(header.DataOffset), io.SeekStart)

    // 5. Load all data blocks
    for i := 0; i < int(header.NumTasks); i++ {
        block, err := readDataBlock(f)
        if err != nil {
            return err
        }
        task := &models.Task{}
        msgpack.Unmarshal(block.Data, task)
        s.tasks[task.ID] = task
    }

    // 6. Replay WAL if present
    if header.Flags&FlagHasWAL != 0 {
        s.replayWAL()
    }

    return nil
}
```

### Strategy 2: Index-Only Load (Fast Startup)

Load only indexes, lazy-load data on demand:

```go
func (s *CustomStorage) LoadIndexesOnly(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()

    // Read header
    header, err := ReadHeader(f)
    if err != nil {
        return err
    }

    // Load only indexes
    f.Seek(int64(header.IndexOffset), io.SeekStart)
    s.loadIndexes(f)

    // Store data section offset for lazy loading
    s.dataOffset = header.DataOffset

    return nil
}

func (s *CustomStorage) GetTask(ctx context.Context, id string) (*models.Task, error) {
    // Check if already loaded
    if task, ok := s.tasks[id]; ok {
        return task, nil
    }

    // Lazy load from disk
    return s.loadTaskFromDisk(id)
}
```

### Strategy 3: Memory-Mapped (Advanced)

Use memory-mapped I/O for zero-copy access:

```go
import "golang.org/x/exp/mmap"

func (s *CustomStorage) LoadMapped(path string) error {
    // Memory-map the file
    m, err := mmap.Open(path)
    if err != nil {
        return err
    }
    s.mmap = m

    // Parse header from mapped memory
    headerBytes := m.At(0)[:HeaderSize]
    header := parseHeader(headerBytes)

    // Indexes accessible without copy
    indexBytes := m.At(int(header.IndexOffset))
    s.parseIndexes(indexBytes)

    return nil
}
```

## Writing Strategies

### Strategy 1: Full Rewrite (Simple)

Rewrite entire file on every save:

```go
func (s *CustomStorage) Save(path string) error {
    s.mu.RLock()
    defer s.mu.RUnlock()

    // Write to temp file
    tempPath := path + ".tmp"
    f, err := os.Create(tempPath)
    if err != nil {
        return err
    }
    defer f.Close()

    // 1. Reserve space for header (write later)
    f.Seek(HeaderSize, io.SeekStart)

    // 2. Write index section
    indexOffset := HeaderSize
    f.Seek(int64(indexOffset), io.SeekStart)
    s.writeIndexes(f)
    indexEnd, _ := f.Seek(0, io.SeekCurrent)

    // 3. Write data section
    dataOffset := indexEnd
    s.writeDataBlocks(f)
    dataEnd, _ := f.Seek(0, io.SeekCurrent)

    // 4. Write header (now that we know offsets)
    header := &FileHeader{
        Magic:         MagicNumber,
        Version:       CurrentVersion,
        CreatedAt:     time.Now().Unix(),
        ModifiedAt:    time.Now().Unix(),
        NumTasks:      uint32(len(s.tasks)),
        NumObjectives: uint32(len(s.objectives)),
        NumCategories: uint32(len(s.categories)),
        IndexOffset:   uint64(indexOffset),
        DataOffset:    uint64(dataOffset),
    }

    // Calculate checksum
    f.Seek(HeaderSize, io.SeekStart)
    h := crc32.NewIEEE()
    io.Copy(h, f)
    header.Checksum = h.Sum32()

    // Write header at beginning
    f.Seek(0, io.SeekStart)
    header.Write(f)

    // Sync to disk
    f.Sync()

    // Atomic rename
    return os.Rename(tempPath, path)
}
```

### Strategy 2: Append + Compaction (Production)

1. **Normal writes**: Append to WAL
2. **Periodic**: Compact WAL into main file

```go
func (s *CustomStorage) CreateTask(ctx context.Context, task *models.Task) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 1. Add to in-memory structures
    s.tasks[task.ID] = task
    s.updateIndexes(task)

    // 2. Append to WAL
    data, _ := msgpack.Marshal(task)
    s.appendToWAL(OpTypeCreate, BlockTypeTask, data)

    // 3. Check if WAL is too large
    if s.walSize() > CompactionThreshold {
        s.scheduleCompaction()
    }

    return nil
}

func (s *CustomStorage) Compact() error {
    // 1. Write full snapshot
    s.Save(s.dataPath)

    // 2. Clear WAL
    os.Truncate(s.walPath, 0)

    return nil
}
```

## Performance Characteristics

### Time Complexity

| Operation | Full Load | Index-Only | Memory-Mapped |
|-----------|-----------|------------|---------------|
| Startup | O(n) | O(i) | O(1) |
| GetTask | O(1) | O(1) + disk | O(1) |
| ListTasks (filtered) | O(k log k) | O(k log k) | O(k log k) |
| Save | O(n) | O(n) | O(n) |

Where:
- n = total records
- i = index size (typically 10-20% of n)
- k = result set size

### Space Complexity

**File Size** (for 10,000 tasks):
```
Header:           256 bytes
Indexes:          ~2 MB (20% of data)
Data:             ~10 MB (1KB per task)
WAL (if active):  ~1-5 MB
Total:            ~12-17 MB
```

**Memory Usage**:
- Full Load: ~12 MB (all in memory)
- Index-Only: ~2 MB (indexes only)
- Memory-Mapped: ~256 bytes (header only, OS manages pages)

### Benchmark Targets

For 10,000 tasks:
- **Load time**: < 100ms (full), < 20ms (index-only), < 1ms (mmap)
- **Query time**: < 1ms
- **Save time**: < 200ms (full rewrite), < 1ms (WAL append)

## Implementation Phases

### Phase 1: Basic Binary Format (Day 1-2)

**Goal**: Read/write binary format with header and data section

**Implementation**:
- Define header structure
- Implement header read/write
- Implement data block read/write
- Use MessagePack for task serialization
- Full rewrite strategy for saves

**Test**: Can save and load tasks

### Phase 2: Add Index Section (Day 3-4)

**Goal**: Store indexes in file

**Implementation**:
- Define index encoding formats
- Write all indexes to index section
- Load indexes on startup
- Verify queries work with loaded indexes

**Test**: Queries as fast as SQLite

### Phase 3: Optimize Loading (Day 5)

**Goal**: Fast startup

**Implementation**:
- Index-only loading mode
- Lazy loading for data on demand
- Benchmark load times

**Test**: Startup < 20ms for 10k tasks

### Phase 4: Add WAL (Day 6-7)

**Goal**: Durability and fast writes

**Implementation**:
- WAL entry format
- Append to WAL on writes
- Replay WAL on startup
- Periodic compaction

**Test**: No data loss on crash, writes < 1ms

### Phase 5: Advanced Features (Day 8+)

**Optional**:
- Compression (gzip, snappy)
- Encryption
- Memory-mapped I/O
- Concurrent reads (RWMutex)
- Incremental compaction

## Testing Strategy

### Unit Tests

```go
func TestCustomStorage(t *testing.T) {
    // Reuse StorageTestSuite
    suite := &storage.StorageTestSuite{
        Factory: func(t *testing.T) storage.Storage {
            path := filepath.Join(t.TempDir(), "test.qdb")
            s, _ := NewCustomStorage(path)
            return s
        },
        Cleanup: func(t *testing.T, s storage.Storage) {
            s.Close()
        },
    }
    suite.RunAllTests(t)
}
```

### Binary Format Tests

```go
func TestHeaderEncoding(t *testing.T) {
    header := &FileHeader{
        Magic:   MagicNumber,
        Version: 1,
        NumTasks: 100,
    }

    buf := new(bytes.Buffer)
    header.Write(buf)

    decoded, _ := ReadHeader(buf)
    assert.Equal(t, header.Magic, decoded.Magic)
    assert.Equal(t, header.NumTasks, decoded.NumTasks)
}

func TestDataBlockEncoding(t *testing.T) {
    task := &models.Task{
        ID:       "test-1",
        Title:    "Test",
        Priority: 3,
    }

    buf := new(bytes.Buffer)
    writeTaskBlock(buf, task)

    block, _ := readDataBlock(buf)
    decodedTask := &models.Task{}
    msgpack.Unmarshal(block.Data, decodedTask)

    assert.Equal(t, task.ID, decodedTask.ID)
    assert.Equal(t, task.Title, decodedTask.Title)
}
```

### Corruption Tests

```go
func TestCorruptionDetection(t *testing.T) {
    // Create valid file
    storage := createStorage()
    storage.CreateTask(ctx, task)
    storage.Save("test.qdb")

    // Corrupt file
    f, _ := os.OpenFile("test.qdb", os.O_RDWR, 0644)
    f.Seek(100, io.SeekStart)
    f.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
    f.Close()

    // Try to load
    storage2 := NewCustomStorage("test.qdb")
    err := storage2.Load()
    assert.Error(t, err)  // Should detect corruption
}
```

## Comparison with Other Formats

| Feature | Custom Binary | SQLite | bbolt | Serialized |
|---------|--------------|--------|-------|------------|
| Setup | High | Medium | Medium | Low |
| Load time | Fast | Medium | Fast | Fast |
| Query speed | Very Fast | Very Fast | Fast | Very Fast |
| Write speed | Medium | Fast | Very Fast | Slow |
| File size | Medium | Medium | Medium | Large |
| Durability | Good (WAL) | Excellent | Excellent | Poor |
| Debug-ability | Hard | Medium | Hard | Easy (JSON) |
| Learning value | High | Low | Medium | Low |

## Recommendations

### Use Custom Binary Format When:
- ✅ Want to learn database internals
- ✅ Need custom optimizations for specific workload
- ✅ Want full control over file format
- ✅ Building production system
- ✅ Dataset size: 10k - 1M records

### Don't Use When:
- ❌ Just prototyping (use serialized)
- ❌ Need SQL queries (use SQLite)
- ❌ Need battle-tested durability (use SQLite/bbolt)
- ❌ Limited implementation time

### Evolution Path:
1. **Prototype**: Start with serialized storage (simple)
2. **Production**: Implement custom binary (optimized)
3. **Scale**: Add compression, WAL, memory-mapping
4. **If outgrow**: Migrate to PostgreSQL

## Tools for Development

### Hex Viewer
```bash
# View binary file structure
hexdump -C data.qdb | head -50

# Check magic number
head -c 4 data.qdb | xxd
```

### File Analyzer
Create a CLI tool to inspect the format:

```bash
# Custom tool
qdb-inspect data.qdb

Output:
Header:
  Magic: QTDO (0x4F445451)
  Version: 1
  Tasks: 10,000
  Index Offset: 256
  Data Offset: 2,048,000

Indexes:
  Status Index: 1,234 bytes
  Priority Index: 567 bytes
  ...

Data Blocks:
  Task blocks: 8,500
  Objective blocks: 12,000
  Category blocks: 10
```

## References

- Storage Interface: `backend/internal/storage/storage.go`
- Storage Test Suite: `backend/internal/storage/testing.go`
- Binary Package: https://pkg.go.dev/encoding/binary
- MessagePack: https://github.com/vmihailenco/msgpack

## Related Documentation

- [Testing Philosophy](../testing/README.md)
- [Storage Layer Tests](../testing/storage-layer-tests.md)
- [SQLite Implementation](./sqlite-implementation.md)
- [Data Models](./data-models-implementation.md)
