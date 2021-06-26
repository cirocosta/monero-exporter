#!/usr/bin/awk -f

# gh-table-printer - prints to `stdout` the metric descriptions
#                    in GitHub flavored markdown tables[1].
#
# [1] - https://help.github.com/articles/organizing-information-with-tables/
#
# Usage: `curl -s localhost:9100/metrics  | grep container_ | ./gh-table-printer`

BEGIN {
  print "| name | description |"
  print "| ---- | ----------- |"
}

/HELP/ {
  line="| " $3 " |"
  for (i = 4; i <= NF; i++) {
    line = line " "$i
  }

  line = line " |"
  print line
}
