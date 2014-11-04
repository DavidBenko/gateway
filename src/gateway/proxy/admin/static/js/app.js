App = Ember.Application.create();

App.Router.map(function() {
  this.resource('admin', function() {
    this.resource('routes');
    this.resource('proxyEndpoints', function() {
      this.resource('newProxyEndpoint', { path: 'new' });
      this.resource('proxyEndpoint', { path: ':endpoint_id' });
    });
    this.resource('libraries', function() {
      this.resource('newLibrary', { path: 'new' });
      this.resource('library', { path: ':library_id' });
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

App.ApplicationAdapter = DS.RESTAdapter.extend({
  namespace: window.location.pathname.replace(/^\//,"").replace(/\/$/,"")
});
