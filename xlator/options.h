#include <stdio.h>
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
} volume_option_t;
