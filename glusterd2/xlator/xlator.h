#include <stdint.h>
#include "options/options.h"

// These definitions are borrowed from libglusterfs/src/xlator.h file in
// glusterfs source. Maintaining this copy here has very little overhead
// right now. Any change to these definitions there should also be updated
// here.


typedef struct xlator_api {
        /* op_version: will be used by volume generation logic to figure
           out whether to insert it in graph or no, based on cluster's
           operating version.
           default value: 0, which means good to insert always */
        uint32_t op_version[GF_MAX_RELEASES];

        /* flags: will be used by volume generation logic to optimize the
           placements etc.
           default value: 0, which means don't treat it specially */
        uint32_t flags;

        /* xlator_id: unique per xlator. make sure to have no collission
           in this ID */
        uint32_t xlator_id;

        /* identifier: a string constant */
        char *identifier;

        /* struct options: if the translator takes any 'options' from the
           volume file, then that should be defined here. optional. */
        volume_option_t *options;
} xlator_api_t;
