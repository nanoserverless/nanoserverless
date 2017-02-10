pi = 0.0
sign = 1.0
n = 1.0
i = 1.0
while i < 100000000:
  pi += (4.0/n)-(4.0/(n+2.0));
  n += 4.0;
  i += 1.0
print pi
