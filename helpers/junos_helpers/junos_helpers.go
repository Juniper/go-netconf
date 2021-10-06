package junos_helpers

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	driver "github.com/davedotdev/go-netconf/drivers/driver"
	sshdriver "github.com/davedotdev/go-netconf/drivers/ssh"

	"golang.org/x/crypto/ssh"
)

const groupStrXML = `<load-configuration action="merge" format="xml">
%s
</load-configuration>
`

const deleteStr = `<edit-config>
	<target>
		<candidate/>
	</target>
	<default-operation>none</default-operation> 
	<config>
		<configuration>
			<groups operation="delete">
				<name>%s</name>
			</groups>
			<apply-groups operation="delete">%s</apply-groups>
		</configuration>
	</config>
</edit-config>`

const commitStr = `<commit/>`

const getGroupStr = `<get-configuration database="committed" format="text" >
  <configuration>
  <groups><name>%s</name></groups>
  </configuration>
</get-configuration>
`

const getGroupXMLStr = `<get-configuration>
  <configuration>
  <groups><name>%s</name></groups>
  </configuration>
</get-configuration>
`

// GoNCClient type for storing data and wrapping functions
type GoNCClient struct {
	Driver driver.Driver
	Lock   sync.RWMutex
}

// Close is a functional thing to close the Driver
func (g *GoNCClient) Close() error {
	g.Driver = nil
	return nil
}

// parseGroupData is a function that cleans up the returned data for generic config groups
func parseGroupData(input string) (reply string, err error) {
	var cfgSlice []string

	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		cfgSlice = append(cfgSlice, scanner.Text())
	}

	var cfgSlice2 []string

	for _, v := range cfgSlice {
		r := strings.NewReplacer("}\\n}\\n", "}", "\\n", "", "/*", "", "*/", "", "</configuration-text>", "")

		d := r.Replace(v)

		// fmt.Println(d)

		cfgSlice2 = append(cfgSlice2, d)
	}

	// Figure out the offset. Each Junos version could give us different stuff, so let's look for the group name
	begin := 0
	end := 0

	for k, v := range cfgSlice2 {
		if v == "groups" {
			begin = k + 4
			break
		}
	}

	// We don't want the very end slice due to config terminations we don't need.
	end = len(cfgSlice2) - 3

	// fmt.Printf("Begin = %v\nEnd = %v\n", begin, end)

	reply = strings.Join(cfgSlice2[begin:end], " ")

	return reply, nil
}

// ReadGroup is a helper function
func (g *GoNCClient) ReadGroup(applygroup string) (string, error) {
	g.Lock.Lock()
	err := g.Driver.Dial()

	if err != nil {
		log.Fatal(err)
	}

	getGroupString := fmt.Sprintf(getGroupStr, applygroup)

	reply, err := g.Driver.SendRaw(getGroupString)
	if err != nil {
		return "", err
	}

	err = g.Driver.Close()

	g.Lock.Unlock()

	if err != nil {
		return "", err
	}

	parsedGroupData, err := parseGroupData(reply.Data)
	return parsedGroupData, nil
}

// UpdateRawConfig deletes group data and replaces it (for Update in TF)
func (g *GoNCClient) UpdateRawConfig(applygroup string, netconfcall string, commit bool) (string, error) {

	deleteString := fmt.Sprintf(deleteStr, applygroup, applygroup)

	g.Lock.Lock()
	err := g.Driver.Dial()
	if err != nil {
		log.Fatal(err)
	}

	reply, err := g.Driver.SendRaw(deleteString)
	if err != nil {
		return "", err
	}

	groupString := fmt.Sprintf(groupStrXML, netconfcall)

	if err != nil {
		log.Fatal(err)
	}

	reply, err = g.Driver.SendRaw(groupString)
	if err != nil {
		return "", err
	}

	if commit == true {
		_, err = g.Driver.SendRaw(commitStr)
		if err != nil {
			return "", err
		}
	}

	err = g.Driver.Close()

	g.Lock.Unlock()

	if err != nil {
		return "", err
	}

	return reply.Data, nil
}

// DeleteConfig is a wrapper for driver.SendRaw()
func (g *GoNCClient) DeleteConfig(applygroup string) (string, error) {

	deleteString := fmt.Sprintf(deleteStr, applygroup, applygroup)

	g.Lock.Lock()
	err := g.Driver.Dial()
	if err != nil {
		log.Fatal(err)
	}

	reply, err := g.Driver.SendRaw(deleteString)
	if err != nil {
		return "", err
	}

	_, err = g.Driver.SendRaw(commitStr)
	if err != nil {
		return "", err
	}

	output := strings.Replace(reply.Data, "\n", "", -1)

	err = g.Driver.Close()

	g.Lock.Unlock()

	if err != nil {
		log.Fatal(err)
	}

	return output, nil
}

// SendCommit is a wrapper for driver.SendRaw()
func (g *GoNCClient) SendCommit() error {
	g.Lock.Lock()

	err := g.Driver.Dial()

	if err != nil {
		return err
	}

	_, err = g.Driver.SendRaw(commitStr)
	if err != nil {
		return err
	}

	g.Lock.Unlock()
	return nil
}

// MarshalGroup accepts a struct of type X and then marshals data onto it
func (g *GoNCClient) MarshalGroup(id string, obj interface{}) error {

	reply, err := g.ReadRawGroup(id)
	if err != nil {
		return err
	}

	err = xml.Unmarshal([]byte(reply), &obj)
	if err != nil {
		return err
	}
	return nil
}

// SendTransaction is a method that unnmarshals the XML, creates the transaction and passes in a commit
func (g *GoNCClient) SendTransaction(id string, obj interface{}, commit bool) error {
	jconfig, err := xml.Marshal(obj)

	if err != nil {
		return err
	}

	// UpdateRawConfig deletes old group by, re-creates it then commits.
	// As far as Junos cares, it's an edit.
	if id != "" {
		_, err = g.UpdateRawConfig(id, string(jconfig), commit)
	} else {
		_, err = g.SendRawConfig(string(jconfig), commit)
	}

	if err != nil {
		return err
	}
	return nil
}

// SendRawConfig is a wrapper for driver.SendRaw()
func (g *GoNCClient) SendRawConfig(netconfcall string, commit bool) (string, error) {

	groupString := fmt.Sprintf(groupStrXML, netconfcall)

	g.Lock.Lock()

	err := g.Driver.Dial()

	if err != nil {
		log.Fatal(err)
	}

	reply, err := g.Driver.SendRaw(groupString)
	if err != nil {
		return "", err
	}

	if commit == true {
		_, err = g.Driver.SendRaw(commitStr)
		if err != nil {
			return "", err
		}
	}

	err = g.Driver.Close()

	if err != nil {
		return "", err
	}

	g.Lock.Unlock()

	return reply.Data, nil
}

// ReadRawGroup is a helper function
func (g *GoNCClient) ReadRawGroup(applygroup string) (string, error) {
	g.Lock.Lock()
	err := g.Driver.Dial()

	if err != nil {
		log.Fatal(err)
	}

	getGroupXMLString := fmt.Sprintf(getGroupXMLStr, applygroup)

	reply, err := g.Driver.SendRaw(getGroupXMLString)
	if err != nil {
		return "", err
	}

	err = g.Driver.Close()

	g.Lock.Unlock()

	if err != nil {
		return "", err
	}

	return reply.Data, nil
}

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

// NewClient returns gonetconf new client driver
func NewClient(username string, password string, sshkey string, address string, port int) (*GoNCClient, error) {

	// Dummy interface var ready for loading from inputs
	var nconf driver.Driver

	d := driver.New(sshdriver.New())

	nc := d.(*sshdriver.DriverSSH)

	nc.Host = address
	nc.Port = port

	// SSH keys takes priority over password based
	if sshkey != "" {
		nc.SSHConfig = &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				publicKeyFile(sshkey),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else {
		// Sort yourself out with SSH. Easiest to do that here.
		nc.SSHConfig = &ssh.ClientConfig{
			User:            username,
			Auth:            []ssh.AuthMethod{ssh.Password(password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}

	nconf = nc

	return &GoNCClient{Driver: nconf}, nil
}

