require 'spec_helper'
require 'app'

describe App do
  include Rack::Test::Methods

  def app
    described_class
  end

  describe 'GET /' do
    before :each do
      get '/'
    end

    it 'is successful' do
      expect(last_response.ok?).to be true
    end

    it 'returns a greeting' do
      expect(last_response.body).to match(/Hello/)
    end
  end
end
