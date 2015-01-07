import Ember from "ember";

export var ProxyEndpointsController = Ember.ArrayController.extend({
  sortProperties: ['name'],
  sortAscending: true
});

export default ProxyEndpointsController;