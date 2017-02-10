pi=0
n=1
i=1
while [ $i -lt 10000 ]; do
  pi=$(bc -l <<< "$pi + (4/$n)-(4/($n+2))")
  let n+=4
  let i++
done
echo $pi
