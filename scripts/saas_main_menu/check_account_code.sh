#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
user="sst_user"
password="demo"
zones=("aws1_saas_mariadb" "aws3_saas_mariadb" "huawei_saas_mariadb" "nbc_saas_sqlserver")

check() {
    cluster="${account_detail[0]}"
    last_updated="${account_detail[1]}"
    zone="${account_detail[2]}"
    echo ""
    echo -e "Zone\t\t: $zone"
    echo -e "Cluster\t\t: $cluster"
    echo -e "Last Updated\t: $last_updated"
}

clear
echo "CHECK ACCOUNT CODE (Other than in NBC SaaS MariaDB)"
echo ""
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
if [[ -z "$account_code" ]]
then
    echo "Please provide a valid account_code";
    exit;
fi
for zone in "${zones[@]}"; do
    eval "${zone}_account_detail=($(mysql -u"$user" -p"$password" -h172.17.70.170 -P3306 -D dbsf_tools -sNe "SELECT vip, DATE(tanggal) FROM dba_all_backup_db_${zone} WHERE dbname LIKE CONCAT('%', '$account_code') AND DATE(tanggal) = CURDATE() - INTERVAL 1 DAY ORDER BY tanggal DESC LIMIT 1;") '$zone')"
done
account_detail=("${aws1_saas_mariadb_account_detail[@]}")
[ ${#account_detail[@]} -eq 1 ] && account_detail=("${aws3_saas_mariadb_account_detail[@]}")
[ ${#account_detail[@]} -eq 1 ] && account_detail=("${huawei_saas_mariadb_account_detail[@]}")
[ ${#account_detail[@]} -eq 1 ] && account_detail=("${nbc_saas_sqlserver_account_detail[@]}")
if [ "${#account_detail[@]}" -gt 1 ]; then
    check
else
    echo "Account code not found"
fi
