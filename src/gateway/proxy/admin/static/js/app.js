App = Ember.Application.create();

App.Router.map(function() {
  this.resource('admin', function() {
    this.resource('routes');
    this.resource('proxyEndpoints', function() {
      this.resource('newProxyEndpoint', { path: 'new' });
      this.resource('proxyEndpoint', { path: ':endpoint_id' });
    });
  });
  this.resource('docs');
  this.resource('support');
});

App.IndexRoute = Ember.Route.extend({
  beforeModel: function() {
    this.transitionTo('admin');
  }
});

App.AdminIndexRoute = Ember.Route.extend({
  beforeModel: function() {
    this.transitionTo('routes');
  }
});

App.ProxyEndpointsIndexRoute = Ember.Route.extend({
  beforeModel: function() {
    this.transitionTo('newProxyEndpoint');
  }
});

App.ApplicationAdapter = DS.RESTAdapter.extend({
  namespace: window.location.pathname.replace(/^\//,"").replace(/\/$/,"")
});
