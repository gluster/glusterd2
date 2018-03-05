#include <stdint.h>

// These definitions are borrowed from libglusterfs/src/options.h file in
// glusterfs source. Maintaining this copy here has very little overhead
// right now. Any change to these definitions there should also be updated
// here.

typedef enum {
        GF_OPTION_TYPE_ANY = 0,
        GF_OPTION_TYPE_STR,
        GF_OPTION_TYPE_INT,
        GF_OPTION_TYPE_SIZET,
        GF_OPTION_TYPE_PERCENT,
        GF_OPTION_TYPE_PERCENT_OR_SIZET,
        GF_OPTION_TYPE_BOOL,
        GF_OPTION_TYPE_XLATOR,
        GF_OPTION_TYPE_PATH,
        GF_OPTION_TYPE_TIME,
        GF_OPTION_TYPE_DOUBLE,
        GF_OPTION_TYPE_INTERNET_ADDRESS,
        GF_OPTION_TYPE_INTERNET_ADDRESS_LIST,
        GF_OPTION_TYPE_PRIORITY_LIST,
        GF_OPTION_TYPE_SIZE_LIST,
        GF_OPTION_TYPE_CLIENT_AUTH_ADDR,
        GF_OPTION_TYPE_MAX,
} volume_option_type_t;

typedef enum {
        GF_OPT_VALIDATE_BOTH = 0,
        GF_OPT_VALIDATE_MIN,
        GF_OPT_VALIDATE_MAX,
} opt_validate_type_t;

typedef enum {
        OPT_FLAG_NONE        = 0,
        OPT_FLAG_SETTABLE    = 1 << 0,
        OPT_FLAG_CLIENT_OPT  = 1 << 1,
        OPT_FLAG_GLOBAL      = 1 << 2,
        OPT_FLAG_FORCE       = 1 << 3,
        OPT_FLAG_NEVER_RESET = 1 << 4,
        OPT_FLAG_DOC         = 1 << 5,
} opt_flags_t;

typedef enum {
        OPT_STATUS_ADVANCED       = 0,
        OPT_STATUS_BASIC          = 1,
        OPT_STATUS_EXPERIMENTAL   = 2,
        OPT_STATUS_DEPRECATED     = 3,
} opt_level_t;

#define GF_MAX_RELEASES 4
#define ZR_VOLUME_MAX_NUM_KEY    4
#define ZR_OPTION_MAX_ARRAY_SIZE 64

typedef struct volume_options {
        char                    *key[ZR_VOLUME_MAX_NUM_KEY];
        volume_option_type_t    otype; // 'type' is a keyword in Go
        double                  min;
        double                  max;
        char                    *value[ZR_OPTION_MAX_ARRAY_SIZE];
        char                    *default_value;
        char                    *description;
        opt_validate_type_t     validate;
        uint32_t                op_version[GF_MAX_RELEASES];
        uint32_t                deprecated[GF_MAX_RELEASES];
        uint32_t                flags;
        char                    *tags[ZR_OPTION_MAX_ARRAY_SIZE];
        char                    *setkey;
        opt_level_t             level;
} volume_option_t;
