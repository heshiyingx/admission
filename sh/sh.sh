#!/usr/bin/env bash
install_cfssl(){
  rm -rf cfss_temp
  mkdir cfss_temp
  cd cfss_temp
  wget https://ghproxy.com/https://github.com/cloudflare/cfssl/releases/download/v1.6.3/cfssl_1.6.3_darwin_amd64
  wget https://ghproxy.com/https://github.com/cloudflare/cfssl/releases/download/v1.6.3/cfssljson_1.6.3_darwin_amd64
  wget https://ghproxy.com/https://github.com/cloudflare/cfssl/releases/download/v1.6.3/cfssl-certinfo_1.6.3_darwin_amd64
  chmod +x cfssl*
  mv cfssl_1.6.3_darwin_amd64 /usr/local/bin/cfssl
  mv cfssljson_1.6.3_darwin_amd64 /usr/local/bin/cfssljson
  mv cfssl-certinfo_1.6.3_darwin_amd64 /usr/local/bin/cfssl-certinfo
  echo "cfssl install success"
  # shellcheck disable=SC2103
  cd ..
  rm -rf cfss_temp
}
gencert(){
  mkdir cert
  cd cert
  if [ -e ca.pem -o -e ca-key.pem ];then
    echo "ca已存在"
    exit 1
  fi

  cat > ca-config.json <<-EOF
  {
    "signing": {
      "default": {
        "expiry": "8760h"
      },
      "profiles": {
        "server": {
          "usages": ["signing", "key encipherment", "server auth", "client auth"],
          "expiry": "8760h"
        }
      }
    }
  }
EOF

  cat > ca-csr.json <<-EOF
  {
      "CN": "kubernetes",
      "key": {
          "algo": "rsa",
          "size": 2048
      },
      "names": [
          {
              "C": "CN",
              "L": "BeiJing",
              "ST": "BeiJing",
              "O": "k8s",
              "OU": "System"
          }
      ]
  }
EOF
  cat > server-csr.json <<-EOF
  {
    "CN": "admission",
    "key": {
      "algo": "rsa",
      "size": 2048
    },
    "names": [
      {
          "C": "CN",
          "L": "BeiJing",
          "ST": "BeiJing",
          "O": "k8s",
          "OU": "System"
      }
    ]
  }
EOF
  cfssl gencert -initca ca-csr.json | cfssljson -bare ca
  cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json \
  		-hostname="image-check.default.svc" -profile=server server-csr.json | cfssljson -bare server
}
cd cert
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json \
  		-hostname="image-check.default.svc" -profile=server server-csr.json | cfssljson -bare server
#gencert
create_secret(){
  kubectl create secret tls admissionpod-tls --cert=./cert/server.pem --key=./goCert/server-key.pem
  cat ./cert/ca.pem | base64
}

