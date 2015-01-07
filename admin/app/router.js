import Ember from 'ember';
import config from './config/environment';

var Router = Ember.Router.extend({
  location: config.locationType
});

Router.map(function() {
  this.resource('admin', function() {
    this.resource('proxyEndpoints', function() {
      this.resource('proxyEndpoints.new', { path: 'new' });
      this.resource('proxyEndpoints.show', { path: ':endpoint_id' });
    });
    this.resource('libraries', function() {
      this.resource('newLibrary', { path: 'new' });
      this.resource('library', { path: ':library_id' });
    });
    this.resource('environments', function() {
      this.resource('newEnvironment', { path: 'new' });
      this.resource('environment', { path: ':environment_id' });
    });
  });
  this.resource('docs');
  this.resource('support');
});

export default Router;
