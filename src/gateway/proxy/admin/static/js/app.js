App = Ember.Application.create();

App.Router.map(function() {
  this.resource('admin', function() {
    this.resource('routes');
    this.resource('endpoints', function() {
      this.resource('newEndpoint', { path: 'new' });
      this.resource('endpoint', { path: ':endpoint_id' });
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

App.AdminController = Ember.ObjectController.extend({
  errorMessage: null,
  successMessage: null,

  actions: {
    closeSuccess: function() {
      this.set('successMessage', null);
    },
    closeError: function() {
      this.set('errorMessage', null);
    }
  }
});


App.ApplicationAdapter = DS.RESTAdapter.extend({
  namespace: window.location.pathname.replace(/^\//,"").replace(/\/$/,"")
});
