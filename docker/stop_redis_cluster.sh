for i in {1..6}
do
  docker rm -f redis-cluster-node-${i}
done
