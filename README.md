# flagconfig
**WARNING: this package unstable and developing slowly.**
## Supported Struct Field Types
flagconfig supports these struct field types:

-[x] string
-[x] uint, uint64
-[ ] uint8, uint16, uint32
-[x] int,  int64
-[ ] int8, int16, int32
-[x] bool
-[x] float64
-[ ] slices of any supported type
-[ ] maps (keys and values of any supported type)
-[x] encoding.TextUnmarshaler
-[x] encoding.BinaryUnmarshaler
-[x] time.Duration
-[x] embedded structs using these fields

Inspired by [envconfig](https://github.com/kelseyhightower/envconfig)