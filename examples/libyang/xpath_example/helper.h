#ifndef HELPER_H
#define HELPER_H

#include <libyang/libyang.h>

typedef void (*clb)(LY_LOG_LEVEL level, const char *msg, const char *path);
void CErrorCallback(LY_LOG_LEVEL level, const char *msg, const char *path);

#endif
