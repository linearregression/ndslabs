// Copyright © 2016 National Data Service
package etcd

import (
	"encoding/json"
	"time"

	api "github.com/ndslabs/apiserver/types"
	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
	"github.com/golang/glog"
)

var etcdBasePath = "/ndslabs/"

type EtcdHelper struct {
	etcd client.KeysAPI
}

func NewEtcdHelper(address string) (*EtcdHelper, error) {

	etcd, err := GetEtcdClient(address)

	return &EtcdHelper{
		etcd: etcd,
	}, err
}

func (s *EtcdHelper) GetAccount(uid string) (*api.Account, error) {
	path := etcdBasePath + "/accounts/" + uid + "/account"

	glog.Infof("GetAccount %s\n", path)

	resp, err := s.etcd.Get(context.Background(), path, nil)

	if err != nil {
		return nil, err
	} else {
		account := api.Account{}
		json.Unmarshal([]byte(resp.Node.Value), &account)
		return &account, nil
	}
}

func (s *EtcdHelper) PutAccount(uid string, account *api.Account) error {

	data, _ := json.Marshal(account)
	opts := client.SetOptions{Dir: true}
	s.etcd.Set(context.Background(), etcdBasePath+"/accounts/"+uid, "", &opts)
	_, err := s.etcd.Set(context.Background(), etcdBasePath+"/accounts/"+uid+"/account", string(data), nil)
	if err != nil {
		glog.Error(err)
		return err
	}

	return nil
}

func (s *EtcdHelper) GetGlobalServices() (*[]api.ServiceSpec, error) {

	services := []api.ServiceSpec{}
	resp, err := s.etcd.Get(context.Background(), etcdBasePath+"/services", nil)
	if err != nil {
		return nil, err
	} else {
		nodes := resp.Node.Nodes
		for _, node := range nodes {
			service := api.ServiceSpec{}
			json.Unmarshal([]byte(node.Value), &service)
			service.Catalog = "system"
			services = append(services, service)
		}
	}
	return &services, nil
}

func (s *EtcdHelper) GetServices(uid string) (*[]api.ServiceSpec, error) {
	services := []api.ServiceSpec{}
	// Get user services
	resp, err := s.etcd.Get(context.Background(), etcdBasePath+"/accounts/"+uid+"/services", nil)
	if err != nil {
		if !client.IsKeyNotFound(err) {
			return nil, err
		}
	} else {
		nodes := resp.Node.Nodes
		for _, node := range nodes {
			service := api.ServiceSpec{}
			json.Unmarshal([]byte(node.Value), &service)
			service.Catalog = "user"
			services = append(services, service)
		}
	}
	return &services, nil
}

func (s *EtcdHelper) GetAllServices(uid string) (*[]api.ServiceSpec, error) {

	services := []api.ServiceSpec{}
	resp, err := s.etcd.Get(context.Background(), etcdBasePath+"/services", nil)
	if err != nil {
		return nil, err
	} else {
		nodes := resp.Node.Nodes
		for _, node := range nodes {
			service := api.ServiceSpec{}
			json.Unmarshal([]byte(node.Value), &service)
			service.Catalog = "system"
			services = append(services, service)
		}
	}

	resp, err = s.etcd.Get(context.Background(), etcdBasePath+"/accounts/"+uid+"/services", nil)
	if err != nil {
		if !client.IsKeyNotFound(err) {
			return nil, err
		}
	} else {
		nodes := resp.Node.Nodes
		for _, node := range nodes {
			service := api.ServiceSpec{}
			json.Unmarshal([]byte(node.Value), &service)
			service.Catalog = "user"
			services = append(services, service)
		}
	}
	return &services, nil
}

func (s *EtcdHelper) PutGlobalService(key string, service *api.ServiceSpec) error {
	data, err := json.Marshal(service)
	if err != nil {
		glog.Error(err)
		return err
	}
	_, err = s.etcd.Set(context.Background(), etcdBasePath+"/services/"+key, string(data), nil)
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

func (s *EtcdHelper) PutService(uid string, key string, service *api.ServiceSpec) error {
	data, err := json.Marshal(service)
	if err != nil {
		glog.Error(err)
		return err
	}
	_, err = s.etcd.Set(context.Background(), etcdBasePath+"/accounts/"+uid+"/services/"+key, string(data), nil)
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

func GetEtcdClient(etcdAddress string) (client.KeysAPI, error) {
	glog.V(3).Infof("GetEtcdClient %s\n", etcdAddress)

	cfg := client.Config{
		Endpoints:               []string{"http://" + etcdAddress},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	kapi := client.NewKeysAPI(c)

	resp, err := kapi.Get(context.Background(), "/", nil)
	_ = resp
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	return kapi, nil
}

func (s *EtcdHelper) GetServiceSpec(uid string, key string) (*api.ServiceSpec, error) {

	// Default to user catalog
	resp, err := s.etcd.Get(context.Background(), etcdBasePath+"/accounts/"+uid+"/services/"+key, nil)
	if err != nil {
		if !client.IsKeyNotFound(err) {
			glog.Error(err)
			return nil, err
		}
	} else {
		service := api.ServiceSpec{}
		node := resp.Node
		json.Unmarshal([]byte(node.Value), &service)
		service.Catalog = "user"
		return &service, nil
	}

	// If not in user catalog, try system catalog
	resp, err = s.etcd.Get(context.Background(), etcdBasePath+"/services/"+key, nil)
	if err != nil {
		if !client.IsKeyNotFound(err) {
			glog.Error(err)
			return nil, err
		}
	} else {
		service := api.ServiceSpec{}
		node := resp.Node
		json.Unmarshal([]byte(node.Value), &service)
		service.Catalog = "system"
		return &service, nil
	}
	return nil, nil
}

func (s *EtcdHelper) GetAccounts() (*[]api.Account, error) {

	glog.V(4).Infoln("GetAccounts()")
	accounts := []api.Account{}

	resp, err := s.etcd.Get(context.Background(), etcdBasePath+"/accounts", nil)

	if err == nil {
		nodes := resp.Node.Nodes
		for _, node := range nodes {

			glog.V(4).Infof("node.Key %s\n", node.Key)
			resp, err = s.etcd.Get(context.Background(), node.Key+"/account", nil)
			if err != nil {
				glog.Error(err)
				return nil, err
			}
			account := api.Account{}
			err := json.Unmarshal([]byte(resp.Node.Value), &account)
			if err != nil {
				glog.Error(err)
				return nil, err
			}
			accounts = append(accounts, account)
		}
	}
	return &accounts, nil
}

func (s *EtcdHelper) GetStack(uid string, sid string) (*api.Stack, error) {

	path := "/accounts/" + uid + "/stacks/" + sid
	resp, err := s.etcd.Get(context.Background(), etcdBasePath+path, nil)

	if err != nil {
		return nil, err
	} else {
		stack := api.Stack{}
		json.Unmarshal([]byte(resp.Node.Value), &stack)
		return &stack, nil
	}
}

func (s *EtcdHelper) PutStack(uid string, sid string, stack *api.Stack) error {
	opts := client.SetOptions{Dir: true}
	s.etcd.Set(context.Background(), etcdBasePath+"/accounts/"+uid, "/stacks", &opts)

	data, _ := json.Marshal(stack)
	path := etcdBasePath + "/accounts/" + uid + "/stacks/" + sid
	//glog.V(4).Infof("stack %s\n", data)
	_, err := s.etcd.Set(context.Background(), path, string(data), nil)
	if err != nil {
		glog.Error("Error storing stack %s", err)
		return err
	} else {
		return nil
	}
}

func (s *EtcdHelper) DeleteAccount(uid string) error {
	_, err := s.etcd.Delete(context.Background(), etcdBasePath+"/accounts/"+uid, &client.DeleteOptions{Recursive: true})
	if err != nil {
		return err
	}
	return nil
}

func (s *EtcdHelper) DeleteGlobalService(key string) error {
	_, err := s.etcd.Delete(context.Background(), etcdBasePath+"/services/"+key, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *EtcdHelper) DeleteService(uid string, key string) error {
	_, err := s.etcd.Delete(context.Background(), etcdBasePath+"/accounts/"+uid+"/services/"+key, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *EtcdHelper) DeleteStack(uid string, sid string) error {
	path := etcdBasePath + "/accounts/" + uid + "/stacks/" + sid
	_, err := s.etcd.Delete(context.Background(), path, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *EtcdHelper) GetStacks(uid string) (*[]api.Stack, error) {

	stacks := []api.Stack{}

	resp, err := s.etcd.Get(context.Background(), etcdBasePath+"/accounts/"+uid+"/stacks", nil)

	if err == nil {
		nodes := resp.Node.Nodes
		for _, node := range nodes {
			stack := api.Stack{}
			err := json.Unmarshal([]byte(node.Value), &stack)
			if err != nil {
				return nil, err
			}
			stacks = append(stacks, stack)
		}
	}
	return &stacks, nil
}

func (s *EtcdHelper) PutVocabulary(name string, vocabulary *api.Vocabulary) error {
	data, err := json.Marshal(vocabulary)
	if err != nil {
		glog.Error(err)
		return err
	}
	_, err = s.etcd.Set(context.Background(), etcdBasePath+"/vocabularies/"+name, string(data), nil)
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

func (s *EtcdHelper) GetVocabulary(name string) (*api.Vocabulary, error) {
	resp, err := s.etcd.Get(context.Background(), etcdBasePath+"/vocabularies/"+name, nil)
	if err != nil {
		if !client.IsKeyNotFound(err) {
			glog.Error(err)
			return nil, err
		}
	} else {
		vocab := api.Vocabulary{}
		json.Unmarshal([]byte(resp.Node.Value), &vocab)
		return &vocab, nil
	}
	return nil, nil
}
