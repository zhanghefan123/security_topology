sudo ./cmc parallel invoke \
--loopNum=1000  \
--threadNum=1000 \
--timeout=30 \
--sleepTime=1000 \
--climbTime=5  \
--use-tls=true \
--check-result=true \
--contract-name="fact" \
--method="save" \
--pairs='[{"key":"data","value":"test_invoke","unique":true}]'  \
--hosts="127.0.0.1:12301" \
--org-IDs="wx-org1.chainmaker.org" \
--user-keys="./testdata/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.key" \
--user-crts="./testdata/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt" \
--org-ids="wx-org1.chainmaker.org" \
--ca-path="./testdata/crypto-config/wx-org1.chainmaker.org/ca" \
--admin-sign-keys="./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key,./testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.key,./testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.key,./testdata/crypto-config/wx-org4.chainmaker.org/user/admin1/admin1.sign.key" \
--admin-sign-crts="./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt,./testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.crt,./testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.crt,./testdata/crypto-config/wx-org4.chainmaker.org/user/admin1/admin1.sign.crt"

#
#sudo ./cmc parallel invoke \
#--hosts="127.0.0.1:12301" \
#--user-crts="./testdata/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt" \
#--user-keys="./testdata/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.key" \
#--org-IDs="wx-org1" \
#--ca-path="./testdata/crypto-config/wx-org1.chainmaker.org/ca" \
#--admin-sign-crts="./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt" \
#--admin-sign-keys="./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key" \
#--chain-id="chain1" \
#--contract-name="fact" \
#--method="save" \
#--pairs='[{"key":"data","value":"test_invoke","unique":true}]' \
#--threadNum=10 \
#--loopNum=1000 \
#--timeout=30 \
#--use-tls=true \
#--check-result=true \
#--output-result=true \
#--record-log=true