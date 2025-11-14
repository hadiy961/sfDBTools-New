mydumper --host 10.100.0.81 --port 3306 --user dbaDO --password 'DataOn24!!' \
  --database dbsf_nbc_dharmagroup \
  --outputdir dbsf_nbc_dharmagroup_backup \
  --compress \
  --threads 16 \
  --chunk-filesize 256 \
  --rows 500000 \
  --statement-size 1000000 \
  --routines \
  --triggers \
  --events \
  --use-savepoints \
  --trx-tables \
  --long-query-guard 3600 \
  --verbose 3 \
  --logfile dbsf_nbc_dharmagroup_backup.log

echo "Backup selesai, memulai restore..."

myloader --host localhost --port 3306 --user root --password 'P@ssw0rdDB' \
  --directory dbsf_nbc_dharmagroup_backup \
  --threads 16 \
  --verbose 3 \
  --drop-tables \
  --overwrite-tables \
  --enable-binlog \
  --max-threads-for-post-actions 4 \
  --serialized-table-creation
