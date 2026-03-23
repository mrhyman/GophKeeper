#!/bin/bash
set -e

CERTS_DIR="./certs"
DAYS=365
COUNTRY="RU"
ORG="GophKeeper"
CN="localhost"

mkdir -p "${CERTS_DIR}"

echo "=== Генерация CA (Certificate Authority) ==="
openssl genrsa -out "${CERTS_DIR}/ca.key" 4096

openssl req -new -x509 \
    -key "${CERTS_DIR}/ca.key" \
    -out "${CERTS_DIR}/ca.crt" \
    -days "${DAYS}" \
    -subj "/C=${COUNTRY}/O=${ORG}/CN=${ORG} CA"

echo "=== Генерация серверного сертификата ==="
openssl genrsa -out "${CERTS_DIR}/server.key" 4096

openssl req -new \
    -key "${CERTS_DIR}/server.key" \
    -out "${CERTS_DIR}/server.csr" \
    -subj "/C=${COUNTRY}/O=${ORG}/CN=${CN}"

cat > "${CERTS_DIR}/server-ext.cnf" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = server
DNS.3 = gophkeeper-server
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

openssl x509 -req \
    -in "${CERTS_DIR}/server.csr" \
    -CA "${CERTS_DIR}/ca.crt" \
    -CAkey "${CERTS_DIR}/ca.key" \
    -CAcreateserial \
    -out "${CERTS_DIR}/server.crt" \
    -days "${DAYS}" \
    -extfile "${CERTS_DIR}/server-ext.cnf"

rm -f "${CERTS_DIR}/server.csr" "${CERTS_DIR}/server-ext.cnf" "${CERTS_DIR}/ca.srl"

chmod 600 "${CERTS_DIR}"/*.key
chmod 644 "${CERTS_DIR}"/*.crt

echo ""
echo "=== Сертификаты сгенерированы ==="
echo "  CA:     ${CERTS_DIR}/ca.crt"
echo "  Server: ${CERTS_DIR}/server.crt"
echo "  Key:    ${CERTS_DIR}/server.key"
echo ""
echo "Для клиента: укажите ca.crt как доверенный CA"
