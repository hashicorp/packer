import Button from '@hashicorp/react-button'
import VERSION from 'data/version'

export default function Homepage() {
  return (
    <div id="p-home">
      <section id="hero">
        <img src="/img/logo-hashicorp.svg" alt="HashiCorp Packer Logo" />
        <h1 className="g-type-display-3">Build Automated Machine Images</h1>
        <div className="buttons">
          <Button
            title="Get Started"
            theme={{ brand: 'packer' }}
            url="https://learn.hashicorp.com/packer"
          />
          <Button
            title={`Download ${VERSION}`}
            theme={{
              variant: 'secondary',
              background: 'light',
            }}
            url="/downloads"
          />
        </div>
      </section>
      <section id="infrastructure-as-code" className="g-container">
        <div className="code-block">
          <div className="circles">
            <span></span>
            <span></span>
            <span></span>
          </div>
          $ packer build template.pkr.hcl
          <span className="green">
            ==&gt; virtualbox: virtualbox output will be in this color.
          </span>
          <span className="blue">
            ==&gt; vmware: vmware output will be in this color.
          </span>
          <span className="blue">
            ==&gt; vmware: Copying or downloading ISO. Progress will be reported
            periodically.
          </span>
          <span className="blue">
            ==&gt; vmware: Creating virtual machine disk
          </span>
          <span className="blue">
            ==&gt; vmware: Building and writing VMX file
          </span>
          <span className="blue">
            ==&gt; vmware: Starting HTTP server on port 8964
          </span>
          <span className="blue">
            ==&gt; vmware: Starting virtual machine...
          </span>
          <span className="green">
            ==&gt; virtualbox: Downloading VirtualBox guest additions. Progress
            will be shown periodically
          </span>
          <span className="green">
            ==&gt; virtualbox: Copying or downloading ISO. Progress will be
            reported periodically.
          </span>
          <span className="green">
            ==&gt; virtualbox: Starting HTTP server on port 8081
          </span>
          <span className="green">
            ==&gt; virtualbox: Creating virtual machine...
          </span>
          <span className="green">
            ==&gt; virtualbox: Creating hard drive...
          </span>
          <span className="green">
            ==&gt; virtualbox: Creating forwarded port mapping for SSH (host
            port 3213)
          </span>
          <span className="green">
            ==&gt; virtualbox: Executing custom VBoxManage commands...
            virtualbox: Executing: modifyvm packer --memory 480 virtualbox:
            Executing: modifyvm packer --cpus
          </span>
          <span className="green">
            ==&gt; virtualbox: Starting the virtual machine...
          </span>
          <span className="blue">==&gt; vmware: Waiting 10s for boot...</span>
          <span className="green">
            ==&gt; virtualbox: Waiting 10s for boot...
          </span>
        </div>
        <div className="text">
          <div className="tag g-type-label">Infrastructure as Code</div>
          <h2 className="g-type-display-2">Modern, Automated</h2>
          <p className="g-type-body">
            HashiCorp Packer automates the creation of any type of machine
            image. It embraces modern configuration management by encouraging
            you to use automated scripts to install and configure the software
            within your Packer-made images. Packer brings machine images into
            the modern age, unlocking untapped potential and opening new
            opportunities.
          </p>
        </div>
      </section>
      <section id="integrations">
        <div className="g-container">
          <div className="logos">
            <img src="/img/integrations/azure.svg" alt="Microsoft Azure Logo" />
            <img
              src="/img/integrations/aws.svg"
              alt="Amazon Web Services Logo"
            />
            <img src="/img/integrations/vmware.svg" alt="VMware Logo" />
            <img
              src="/img/integrations/google-cloud.svg"
              alt="Google Cloud Platform Logo"
            />
            <img src="/img/integrations/docker.svg" alt="Docker Logo" />
            <img
              src="/img/integrations/digitalocean.svg"
              alt="DigitalOcean Logo"
            />
          </div>
          <div className="text">
            <div className="tag g-type-label">Integrations</div>
            <h2 className="g-type-display-2">Works Out of The Box</h2>
            <p className="g-type-body">
              Out of the box Packer comes with support to build images for
              Amazon EC2, CloudStack, DigitalOcean, Docker, Google Compute
              Engine, Microsoft Azure, QEMU, VirtualBox, VMware, and more.
              Support for more platforms is on the way, and anyone can add new
              platforms via plugins.
            </p>
          </div>
        </div>
      </section>
    </div>
  )
}
