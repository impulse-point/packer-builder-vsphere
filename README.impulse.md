Forked from <https://github.com/jetbrains-infra/packer-builder-vsphere>.

Modified `common/step_run.go` to avoid setting boot order at all. Doing so as part of creating an image via `vsphere-iso` results in an OVF file containing a section such as:
```
<vmw:BootOrderSection vmw:instanceId="8" vmw:type="disk">
	<Info>Virtual hardware device boot order</Info>
</vmw:BootOrderSection>
```

This section causes some vCenter/vSphere environments to fail to import the image.

Building the plugin:
```
docker-compose run build
```