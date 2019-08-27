cluster_config="--cluster-enabled yes \
  --cluster-require-full-coverage no \
  --cluster-node-timeout 15000 \
  --cluster-migration-barrier 1 \
  --protected-mode no"

for i in {1..6}
do
  port=$((13000+$i))
  data_dir=$(pwd)/cluster-data/data-${i}
  mkdir -p ${data_dir}
  nohup redis-server --port ${port} --dir ${data_dir} --cluster-config-file ${data_dir}/nodes.conf $cluster_config &
done
