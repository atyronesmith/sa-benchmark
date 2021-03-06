#ifndef _UTILS_H
#define _UTILS_H 1

#define _GNU_SOURCE
#include <features.h>

#include <time.h>
#include <stdio.h>
#include <inttypes.h>

void time_diff(struct timespec t1, struct timespec t2, struct timespec *diff);
char *time_sprintf(char *buf, struct timespec t1);
int64_t now();

#endif