package control

import (
	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
)

var yamlDatacontrol_sts = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: control
spec:
  selector:
    matchLabels:
      app: control
  serviceName: "control"
  replicas: 1
  template:
    metadata:
      labels:
        app: control
        contrail_manager: control
    spec:
      securityContext:
        #runAsGroup: 1999
        fsGroup: 1999
      containers:
      - image: docker.io/michaelhenkel/contrail-controller-control-control:5.2.0-dev1
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        imagePullPolicy: Always
        name: control
        volumeMounts:
        - mountPath: /var/log/contrail
          name: control-logs
      - image: docker.io/michaelhenkel/contrail-controller-control-dns:5.2.0-dev1
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        imagePullPolicy: Always
        name: dns
        volumeMounts:
        - mountPath: /var/log/contrail
          name: control-logs
        - mountPath: /etc/contrail
          name: etc-contrail
        - mountPath: /etc/contrail/dns
          name: etc-contrail-dns
      - image: docker.io/michaelhenkel/contrail-controller-control-named:5.2.0-dev1
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        imagePullPolicy: Always
        name: named
        securityContext:
          privileged: true
          runAsGroup: 1999
        volumeMounts:
        - mountPath: /var/log/contrail
          name: control-logs
        - mountPath: /etc/contrail
          name: etc-contrail
        - mountPath: /etc/contrail/dns
          name: etc-contrail-dns
      - env:
        - name: NODE_TYPE
          value: control
        - name: DOCKER_HOST
          value: unix://mnt/docker.sock
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        image: docker.io/michaelhenkel/contrail-nodemgr:5.2.0-dev1
        imagePullPolicy: Always
        name: nodemanager
        lifecycle:
          preStop:
            exec:
              command:
              - python /etc/mycontrail/deprovision.sh.${POD_IP}
        volumeMounts:
        - mountPath: /var/log/contrail
          name: control-logs
        - mountPath: /var/crashes
          name: crashes
        - mountPath: /mnt
          name: docker-unix-socket
      dnsPolicy: ClusterFirst
      hostNetwork: true
      initContainers:
      - command:
        - sh
        - -c
        - until grep ready /tmp/podinfo/pod_labels > /dev/null 2>&1; do sleep 1; done
        env:
        - name: CONTRAIL_STATUS_IMAGE
          value: docker.io/michaelhenkel/contrail-status:5.2.0-dev1
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        image: busybox
        imagePullPolicy: Always
        name: init
        volumeMounts:
        - mountPath: /tmp/podinfo
          name: status
      - env:
        - name: CONTRAIL_STATUS_IMAGE
          value: docker.io/michaelhenkel/contrail-status:5.2.0-dev1
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        image: docker.io/michaelhenkel/contrail-node-init:5.2.0-dev1
        imagePullPolicy: Always
        name: nodeinit
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /host/usr/bin
          name: host-usr-bin
      nodeSelector:
        node-role.kubernetes.io/master: ""
      restartPolicy: Always
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
      - hostPath:
          path: /var/log/contrail/control
          type: ""
        name: control-logs
      - hostPath:
          path: /var/log/contrail/control
          type: ""
        name: crashes
      - hostPath:
          path: /var/run
          type: ""
        name: docker-unix-socket
      - hostPath:
          path: /usr/bin
          type: ""
        name: host-usr-bin
      - emptyDir: {}
        name: etc-contrail
      - emptyDir: {}
        name: etc-contrail-dns
      - downwardAPI:
          defaultMode: 420
          items:
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.labels
            path: pod_labels
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.labels
            path: pod_labelsx
        name: status`

func GetSTS() *appsv1.StatefulSet {
	sts := appsv1.StatefulSet{}
	err := yaml.Unmarshal([]byte(yamlDatacontrol_sts), &sts)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDatacontrol_sts))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &sts)
	if err != nil {
		panic(err)
	}
	return &sts
}
