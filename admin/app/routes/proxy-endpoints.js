import Ember from "ember";

export var ProxyEndpoints = Ember.Route.extend({
  controllerName: "proxyEndpoints/index",
  model: function() {
    return [];
    // return this.store.find('proxyEndpoint');
  }
});

export var ProxyEndpointsNew = Ember.Route.extend({
  controllerName: "proxyEndpoints/new",
  templateName: 'proxyEndpoints/new',
  model: function() {
    return this.store.createRecord('proxyEndpoint');
  }
});

export default ProxyEndpoints;
